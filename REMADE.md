# Test Flow 实现指南

## 功能概述

本实现支持 `test_flow.yml` 中定义的高级步骤执行逻辑，包括：

1. **步骤ID标识**: 每个步骤可以通过 `id` 字段定义唯一标识
2. **响应字段追加**: 通过 `response` 字段可以向API返回值中追加自定义字段
3. **请求变量定义**: 通过 `request` 字段可以定义请求中的变量，用于替换API中的 `{{}}` 占位符
4. **跨步骤变量引用**: 支持引用其他步骤的返回值，如 `sendTBlockC.response.receipt`

## YAML 格式说明

```yaml
cases:
  - name: "测试用例名称"
    steps:
      API方法名:
        id: "步骤唯一标识"  # 可选
        request:  # 可选，定义请求变量
          变量名: 值或引用
        response:  # 可选，追加到返回值的字段
          字段名: 值
    loop: 循环次数
    thread: 并发线程数
    variables:
      全局变量名: 值
```

## 示例

```yaml
cases:
  - name: "部署合约"
    steps:
      wallet_sendTBlockC:
        id: "sendTBlockC"
        response:
          var1: "1234"
      wallet_getContractState:
        request:
          contractAddr: "sendTBlockC.response.receipt"
    loop: 10
    thread: 10
    variables:
      code: "0x608060405..."
      account: "zltc_XzRRPepmmCSF8734cFNp89tg6GHo17bcU"
      password: "123456"
```

## 变量引用规则

1. **全局变量**: 直接使用变量名，如 `{{account}}`
2. **跨步骤引用**: 使用点号分隔，格式为 `步骤ID.response.字段名`
   - 例如: `sendTBlockC.response.receipt` 表示获取ID为 `sendTBlockC` 的步骤返回值中的 `receipt` 字段

## 执行流程

1. 加载测试流程配置
2. 按顺序执行每个步骤
3. 存储带有ID的步骤返回值
4. 解析跨步骤变量引用
5. 合并response字段到返回值
6. 传递给下一个步骤

## 使用方法

```bash
# 使用默认配置文件
./perf-tester.exe

# 指定自定义配置文件
./perf-tester.exe -flow test_flow_example.yml -config config.yml -api api.json
```