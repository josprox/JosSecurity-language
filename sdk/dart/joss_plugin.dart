import 'dart:async';
import 'dart:convert';
import 'dart:io';

typedef JossMethod = FutureOr<dynamic> Function(List<dynamic> args);

/// Robust Joss JP v2 Dart/Flutter Sidecar Framework (joss-rpc-v1).
class JossPlugin {
  static const String protocol = 'joss-rpc-v1';
  final Map<String, JossMethod> _methods = {};

  void register(String name, JossMethod handler) {
    _methods[name] = handler;
  }

  Future<void> run() async {
    var requestId = '';
    try {
      final text = await stdin.transform(utf8.decoder).join();
      if (text.trim().isEmpty) {
        throw StateError('Stdin received empty request');
      }

      final request = jsonDecode(text) as Map<String, dynamic>;
      requestId = request['id']?.toString() ?? '';

      if (request['protocol'] != protocol) {
        throw StateError('Unsupported protocol version: ${request['protocol']}');
      }

      final method = request['method']?.toString() ?? '';
      final handler = _methods[method];
      if (handler == null) {
        throw ArgumentError('Unknown method: $method');
      }

      final args = (request['args'] as List<dynamic>?) ?? const [];
      final result = await handler(args);

      if (result is Stream) {
        await for (final chunk in result) {
          _writeFrame({'id': requestId, 'event': 'chunk', 'content': chunk});
        }
        _writeFrame({'id': requestId, 'result': null});
        return;
      }

      _writeFrame({'id': requestId, 'result': result});
    } catch (error, stack) {
      _writeFrame({
        'id': requestId,
        'error': {
          'code': error.runtimeType.toString(),
          'message': error.toString(),
        }
      });
    }
  }

  void _writeFrame(Map<String, dynamic> frame) {
    stdout.writeln(jsonEncode(frame));
  }
}

/// Legacy entry point helper.
Future<void> runJossPlugin(Map<String, JossMethod> methods) async {
  final plugin = JossPlugin();
  methods.forEach((name, handler) => plugin.register(name, handler));
  await plugin.run();
}
