# 订单状态机

```text
CREATED
  -> PAYMENT_PENDING
  -> PAYMENT_CONFIRMED
  -> DELIVERY_PREPARING
  -> DELIVERY_READY
  -> TRANSFERRING
  -> DELIVERED
  -> COMPLETED
```

终止和争议状态：`CANCELLED`、`DISPUTED`、`REFUND_PENDING`、`REFUNDED`。

## 关键约束

- 创建订单时保存数据版本、价格、佣金比例和许可协议快照。
- 进入 `PAYMENT_PENDING` 前必须成功冻结预计佣金。
- 支付失败或订单取消时释放冻结佣金。
- 只有 `PAYMENT_CONFIRMED` 之后才能创建 Delivery Grant。
- Controller 不得直接更新订单状态，所有转换由领域服务执行并写入 `order_events`。
