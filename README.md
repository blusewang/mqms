# MessageQueue MicroServices Framework
一款基于消息队列的消息驱动型微服务框架。支持链路追踪、可靠延迟消息、错误收集、失败重放。

## 设计目标
适用于稳健要求高的事件驱动式的业务场景。比如：
- 收到用户付款成功事件后，要处理的与购买相对应的一批业务（标记订单、处理用户虚拟资产、发推送消息、发通知、延时检查等子任务）。
- 收到微信公众号事件后，要处理的一批自定义的本地化业务。
- 收到平台api处理过程中发出的事件后，要处理的一批自定义的本地化业务。

