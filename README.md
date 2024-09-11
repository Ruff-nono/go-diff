
```mermaid
flowchart LR
A[requests] -->|Http| B(Diff)
B -->|Http| D[Target 1]
B -->|Http| E[Target 2]
```
![img.png](img.png)

# 什么是Diff工具
Diff 用于充当代理，将收到的任何请求，发送到两个正在运行的实例。然后，它会比较响应。并比较两个响应不同的部分。

# 如何使用
1、构建并启动
```shell
go build .

./go-diff &
```
2、go-diff 默认端口为18080，将请求流量转发至go-diff，go-diff则会将请求同时转发到配置文件中的host1和host2，并对比响应体输出到日志`diff.log`

```
time: 2024/06/17 10:30:07 start at 2024-06-17 10:30:07
time: 2024/06/17 10:30:07 Starting proxy server, forwarding requests to: http://localhost:8080 and http://localhost:8081
time: 2024/06/17 10:36:44 [diff] Request: /api/register
error: Status codes are different. First: 401, Second: 400
time: 2024/06/17 10:36:50 [diff] Request: /api/register
error: Response bodies are different.
First response: {"code":"C101","msg":"鉴权数据错误 decode code error:invalid timestamp"}
Second response: {"code":"C001","msg":"decrypt code failed"}
```
# 配置文件
默认为同路径下的`config.json`
```json
{
  "host1": "http://localhost:8081", // 目标地址A
  "host2": "http://localhost:8081", // 目标地址B
  "headers_include": [  // 表示要比对的response header
    "Content-Type",
    "Content-Length"
  ],
  "compare_status_code": true,  // 是否比对response status
  "equivalent_status_codes": [  // 哪些status 可以被认为是相同含义并忽略，如示例 400与401可看作相同含义
    [
      400,
      401
    ]
  ],
  "compare_body": false,  // 是否比对response body
  "bodies_exclude": [   // 针对 json body， 哪些字段可以被忽略比对
    "/secret",
    "/uniqueId"
  ]
}

```
json body的字段填写规则
```json
{
  "a": { "b": { "c": { "x": 1, "y": 2, "z": 3 } } }
}
```
/a/b/c/x
/a/b/c/y
/a/b/c/z

更多详细规则，详见引用[json diff](https://github.com/wI2L/jsondiff)

配合[`go replay`](https://github.com/buger/goreplay)工具食用更佳
