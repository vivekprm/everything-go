package decoupling

import (
	_ "github.com/go-sql-driver/mysql"
)

type CustomerService struct {
	store mysql.Store // Depends on the concrete implementation
}

func (cs CustomerService) CreateNewCustomer(id string) error {
	customer := Customer{id: id}
	return cs.store.StoreCustomer(customer)
}
