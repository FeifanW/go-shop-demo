package repositories

import (
	"database/sql"
	"go-shop-demo/common"
	"go-shop-demo/datamodels"
	"strconv"
)

type IOrderRepository interface {
	Conn() error
	Insert(*datamodels.Order) (int64, error)      // 增
	Delete(int64) bool                            // 删
	Update(*datamodels.Order) error               // 改
	SelectByKey(int64) (*datamodels.Order, error) // 查
	SelectAll() ([]*datamodels.Order, error)      //
	SelectAllWithInfo() (map[int]map[string]string, error)
}

// Go语言没有构造函数
func NewOrderMangerRepository(table string, sql *sql.DB) IOrderRepository {
	return &OrderMangerRepository{table, sql}
}

type OrderMangerRepository struct {
	table     string
	mysqlConn *sql.DB
}

// 连接函数
func (o *OrderMangerRepository) Conn() error {
	if o.mysqlConn == nil {
		mysql, err := common.NewMysqlConn()
		if err != nil {
			return err
		}
		o.mysqlConn = mysql
	}
	if o.table == "" {
		o.table = "order"
	}
	return nil
}

// 插入函数
func (o *OrderMangerRepository) Insert(order *datamodels.Order) (productID int64, err error) {
	if err = o.Conn(); err != nil {
		return
	}
	sql := "INSERT" + o.table + "set userID=?,productID=?,orderStatus=?"
	stmt, errStmt := o.mysqlConn.Prepare(sql)
	if errStmt != nil {
		return productID, err
	}
	result, errResult := stmt.Exec(order.UserId, order.ProductId, order.OrderStatus)
	if errResult != nil {
		return productID, errResult
	}
	return result.LastInsertId()
}

// 删除
func (o *OrderMangerRepository) Delete(orderID int64) (isOK bool) {
	if err = o.Conn(); err != nil {
		return
	}
	sql := "delete from" + o.table + "where ID = ?"
	stmt, errStmt := o.mysqlConn.Prepare(sql)
	if errStmt != nil {
		return false
	}
	_, err := stmt.Exec(orderID)
	if err != nil {
		return false
	}
	return true
}

// 更新
func (o *OrderMangerRepository) Update(order *datamodels.Order) (err error) {
	if errConn := o.Conn(); errConn != nil {
		return errConn
	}
	sql := "Update" + o.table + "set userID=?,productID=?,orderStatus=? Where ID=" + strconv.FormatInt(order.ID, 10)
	stmt, errStmt := o.mysqlConn.Prepare(sql)
	if errStmt != nil {
		return errStmt
	}
	_, err = stmt.Exec(order.UserId, order.ProductId, order.OrderStatus)
	return
}

// 查询，根据订单ID查询结构体
func (o *OrderMangerRepository) SelectByKey(orderID int64) (order *datamodels.Order, err error) {
	if errConn := o.Conn(); errConn != nil {
		return &datamodels.Order{}, errConn
	}
	sql := "Select * From" + o.table + "where ID=" + strconv.FormatInt(orderID, 10)
	row, errRow := o.mysqlConn.Query(sql)
	if errRow != nil {
		return &datamodels.Order{}, errRow
	}
	result := common.GetResultRow(row)
	if len(result) == 0 {
		return &datamodels.Order{}, err
	}
	order = &datamodels.Order{}
	common.DataToStructByTagSql(result, order)
	return
}
