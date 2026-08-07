use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::collections::HashMap;
use std::io::{self, Read};
use std::panic::{catch_unwind, AssertUnwindSafe};

pub const PROTOCOL: &str = "joss-rpc-v1";

pub type MethodHandler = Box<dyn Fn(Vec<Value>) -> Result<Value, String> + Send + Sync>;

#[derive(Deserialize)]
struct Request {
    protocol: String,
    id: String,
    method: String,
    #[serde(default)]
    args: Vec<Value>,
}

#[derive(Serialize)]
struct RpcError {
    code: String,
    message: String,
}

#[derive(Serialize)]
struct Response {
    id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    result: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<RpcError>,
}

pub struct JossPlugin {
    methods: HashMap<String, MethodHandler>,
}

impl JossPlugin {
    pub fn new() -> Self {
        Self {
            methods: HashMap::new(),
        }
    }

    pub fn register<F>(mut self, name: &str, handler: F) -> Self
    where
        F: Fn(Vec<Value>) -> Result<Value, String> + Send + Sync + 'static,
    {
        self.methods.insert(name.to_string(), Box::new(handler));
        self
    }

    pub fn run(self) -> Result<(), Box<dyn std::error::Error>> {
        let mut input = String::new();
        io::stdin().read_to_string(&mut input)?;

        if input.trim().is_empty() {
            let resp = Response {
                id: "".into(),
                result: None,
                error: Some(RpcError {
                    code: "EMPTY_REQUEST".into(),
                    message: "Stdin received no data".into(),
                }),
            };
            println!("{}", serde_json::to_string(&resp)?);
            return Ok(());
        }

        let request: Request = match serde_json::from_str(&input) {
            Ok(req) => req,
            Err(e) => {
                let resp = Response {
                    id: "".into(),
                    result: None,
                    error: Some(RpcError {
                        code: "PARSE_ERROR".into(),
                        message: e.to_string(),
                    }),
                };
                println!("{}", serde_json::to_string(&resp)?);
                return Ok(());
            }
        };

        let request_id = request.id.clone();

        if request.protocol != PROTOCOL {
            let resp = Response {
                id: request_id,
                result: None,
                error: Some(RpcError {
                    code: "UNSUPPORTED_PROTOCOL".into(),
                    message: format!("Unsupported protocol version: {}", request.protocol),
                }),
            };
            println!("{}", serde_json::to_string(&resp)?);
            return Ok(());
        }

        let handler = match self.methods.get(&request.method) {
            Some(h) => h,
            None => {
                let resp = Response {
                    id: request_id,
                    result: None,
                    error: Some(RpcError {
                        code: "UNKNOWN_METHOD".into(),
                        message: format!("Unknown method: {}", request.method),
                    }),
                };
                println!("{}", serde_json::to_string(&resp)?);
                return Ok(());
            }
        };

        // Execute handler wrapped in catch_unwind for safety against panics
        let result = catch_unwind(AssertUnwindSafe(|| handler(request.args)));

        let response = match result {
            Ok(Ok(val)) => Response {
                id: request_id,
                result: Some(val),
                error: None,
            },
            Ok(Err(msg)) => Response {
                id: request_id,
                result: None,
                error: Some(RpcError {
                    code: "PLUGIN_ERROR".into(),
                    message: msg,
                }),
            },
            Err(_) => Response {
                id: request_id,
                result: None,
                error: Some(RpcError {
                    code: "PANIC_ERROR".into(),
                    message: "Plugin execution panicked".into(),
                }),
            },
        };

        println!("{}", serde_json::to_string(&response)?);
        Ok(())
    }
}

/// Legacy entry point compatibility
pub fn run(methods: HashMap<String, Box<dyn Fn(Vec<Value>) -> Result<Value, String> + Send + Sync>>) -> Result<(), Box<dyn std::error::Error>> {
    let mut plugin = JossPlugin::new();
    plugin.methods = methods;
    plugin.run()
}
