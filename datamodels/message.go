package datamodels

// 简单的结构体
type Message struct {
	ProductID int64
	UserID    int64
}

// 创建结构体
func NewMessage(userId int64, productId int64) *Message {
	return &Message{UserID: productId, ProductID: productId}
}
