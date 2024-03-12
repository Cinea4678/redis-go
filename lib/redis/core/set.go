package core

const (
	encIntset = iota // 底层类型：整数集合
	encDict          // 底层类型：Dict
)

// Set
/**
根据存储内容自动选择底层的数据结构；
在存储少量整数（32个及以下）时，采用整数集合；否则，采取哈希表
*/
type Set struct {
	enc byte
	ptr interface{}
}
