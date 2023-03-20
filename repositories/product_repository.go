package repositories

import (
	"database/sql"
	"go-shop-demo/common"
	"go-shop-demo/datamodels"
)

// 第一步，先开发对应的接口
// 第二步，实现定义的接口

type IProduct interface {
	// 连接数据
	Conn() error
	Insert(*datamodels.Product) (int64, error)
	Delete(int64) bool
	Update(*datamodels.Product) error
	SelectByKey(int64) (*datamodels.Product, error)
	SelectAll() ([]*datamodels.Product, error)
}

// 实现接口
type ProductManager struct {
	table     string
	mysqlConn *sql.DB
}

func NewProductManager(table string, db *sql.DB) IProduct {
	return &ProductManager{table, db}
}

// 数据连接
func (p *ProductManager) Conn() (err error) {
	if p.mysqlConn == nil {
		mysql, err := common.NewMysqlConn() // 创建连接
		if err != nil {
			return err
		}
		p.mysqlConn = mysql
	}
	if p.table == "" {
		p.table = "product"
	}
	return
}

// 插入
func (p *ProductManager) Insert(product *datamodels.Product) (productId int64, err error) {
	// 1.判断连接是否存在
	if err = p.Conn(); err != nil {
		return
	}
	// 2.准备sql
	sql := "INSERT product SET productName=?,productNum=?,productImage=?,productUrl=?"
	stmt, errSql := p.mysqlConn.Prepare(sql)
	if errSql != nil {
		return 0, errSql
	}
	// 3.传入参数
	result, errStmt := stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl)
	if errStmt != nil {
		return 0, errStmt
	}
	productId, err = result.LastInsertId()
	return
}
