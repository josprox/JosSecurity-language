# Baseline del runtime: 9d27239

- Fecha: 2026-08-31
- Host: Intel Core i5-10300H, Windows/amd64
- Go: 1.26.2
- Muestras: 5
- Benchtime: 100 ms por muestra
- Métrica mostrada: mediana observada

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| Startup/Parse | 26,698 | 9,124 | 183 |
| Startup/Analyze | 8,352 | 5,938 | 37 |
| Startup/LoadSerializedAST | 255,028 | 80,208 | 857 |
| Startup/FirstExecution | 2,469 | 2,080 | 16 |
| Startup/RepeatedExecution | 1,747 | 760 | 7 |
| Basic/Assignment | 140.6 | 0 | 0 |
| Basic/VariableLocalMap | 26.71 | 0 | 0 |
| Basic/ArithmeticInt | 53.94 | 0 | 0 |
| Basic/Comparison | 54.47 | 0 | 0 |
| Basic/BooleanShortCircuit | 81.28 | 0 | 0 |
| Basic/StringConversion | 53.30 | 16 | 1 |
| Basic/ArrayRead literal | 289.1 | 88 | 2 |
| Basic/MapRead literal | 793.3 | 400 | 6 |
| Basic/StringIndexASCII | 173.3 | 36 | 3 |
| Basic/StringIndexUnicode | 155.2 | 36 | 3 |
| Basic/VariableGlobalMap | 26.76 | 0 | 0 |
| Basic/TypedStringToInt | 104.5 | 8 | 1 |
| Control/LoopSmall | 4,098 | 760 | 7 |
| Control/LoopLarge | 2,099,357 | 78,720 | 9,752 |
| Control/Match | 2,217 | 888 | 11 |
| Control/TernaryBlock | 1,843 | 760 | 7 |
| Control/Recursion | 22,500 | 7,784 | 84 |
| Functions/Simple | 1,770 | 760 | 7 |
| Functions/Nested | 5,717 | 2,312 | 23 |
| Functions/Ref | 2,145 | 767 | 7 |
| Functions/Closure | 1,855 | 824 | 8 |
| Objects/Create | 933.8 | 456 | 7 |
| Objects/PropertyRead | 181.3 | 0 | 0 |
| Objects/PropertyWrite | 401.7 | 32 | 2 |
| Objects/InheritedPropertyRead | 250.7 | 0 | 0 |
| Objects/MethodCall | 2,201 | 504 | 7 |
| Objects/InheritedMethodResolution | 2,817 | 512 | 8 |
| Objects/StaticMethod | 2,335 | 880 | 11 |
| Collections/ArrayConstruct | 415.7 | 216 | 3 |
| Collections/ArrayRead | 46.15 | 0 | 0 |
| Collections/ArrayWrite | 60.59 | 0 | 0 |
| Collections/ArrayGrowth | 431.5 | 240 | 4 |
| Collections/MapConstruct | 689.4 | 400 | 6 |
| Collections/MapRead | 114.7 | 16 | 1 |
| Collections/MapWrite | 180.0 | 32 | 2 |
| Collections/Iteration | 1,157 | 0 | 0 |
| Runtime/Exception | 1,738 | 160 | 4 |
| Runtime/AsyncAwait | 4,166 | 1,680 | 18 |
| Runtime/ChannelRoundTrip | 291.7 | 136 | 3 |
| Apps/JSONProcessing | 3,635 | 1,464 | 27 |
| Apps/CRUDMapping | 2,653 | 1,160 | 13 |
| Apps/HTTPHandler | 2,665 | 1,144 | 12 |
| Apps/ArrayTransform | 5,366 | 1,216 | 20 |
| Apps/TemplateRendering | 3,666 | 968 | 23 |
| Apps/DBMapping | 3,691 | 1,016 | 17 |

El comando completo pasó. El perfil con `memprofilerate=1` se usó únicamente
para atribución de allocations y no para estas cifras de velocidad.
