# BlctekIP Dataset Manifest v1

## 目标

确保同一数据目录在受支持的平台上生成相同的文件清单和 Merkle Root。

## 规范

| 项目 | 规则 |
|---|---|
| 哈希算法 | SHA-256 |
| 分块大小 | 4 MiB，最后一块可小于该值 |
| 路径编码 | UTF-8 |
| 路径分隔符 | `/` |
| 文件排序 | 规范化相对路径按 UTF-8 字节升序 |
| 符号链接 | 拒绝 |
| 隐藏文件 | 默认包含，可由显式排除规则排除 |
| 时间与权限 | 不参与内容哈希 |
| 空目录 | 不参与根哈希 |

## 文件叶子

```text
file_leaf = SHA256(
  "BLCTEKIP_FILE_V1" ||
  uint64_be(path_length) || normalized_path ||
  uint64_be(file_size) ||
  uint32_be(chunk_count) || ordered_chunk_hashes
)
```

## Merkle 节点

```text
parent = SHA256("BLCTEKIP_NODE_V1" || left || right)
```

当某一层节点数量为奇数时，最后一个节点复制后与自身组合。空数据集不允许发布。

任何参与 Manifest 的路径、文件内容、排除规则或分块规范发生变化，都必须创建新的 `dataset_version`。
