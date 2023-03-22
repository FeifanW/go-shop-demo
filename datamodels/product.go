package datamodels

type Product struct {
	ID           int64  `json:"id" sql:"ID" imooc:"ID"`
	ProductName  string `json:"ProductName" sql:"ProductName" imooc:"ProductName"`
	ProductNum   int64  `json:"Product_num" sql:"ProductNum" imooc:"ProductNum" `
	ProductImage string `json:"ProductImage" sql:"ProductImage" imooc:"ProductImage"`
	ProductUrl   string `json:"Product_url" sql:"ProductUrl" imooc:"ProductUrl"`
}
