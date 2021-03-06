package protocol

import (
	"errors"
)

type Context struct {
	Account
	changes []Change
	FundsTx
}

type Change struct {
	index int
	value []byte
}

func NewChange(index int, value []byte) Change {
	return Change{index, value}
}

func (c *Change) GetChange() (int, []byte) {
	return c.index, c.value
}

func NewContext(account Account, fundsTx FundsTx) *Context {
	newContext := Context{
		Account: account,
		changes: []Change{},
		FundsTx: fundsTx,
	}
	return &newContext
}

func (c *Context) GetContract() []byte {
	return c.Contract
}

func (c *Context) GetContractVariable(index int) ([]byte, error) {
	if index >= len(c.ContractVariables) || index < 0 {
		return []byte{}, errors.New("Index out of bounds")
	}
	variable := []byte(c.ContractVariables[index])
	cp := make([]byte, len(variable))
	copy(cp, variable)

	return cp, nil
}

func (c *Context) SetContractVariable(index int, value []byte) error {
	if len(c.ContractVariables) <= index {
		return errors.New("Index out of bounds")
	}

	cp := make([]byte, len(value))
	copy(cp, value)

	change := NewChange(index, cp)
	c.changes = append(c.changes, change)
	return nil
}

func (c *Context) PersistChanges() {
	for _, change := range c.changes {
		i, value := change.GetChange()
		c.ContractVariables[i] = value
	}
}

func (c *Context) GetAddress() [64]byte {
	return c.Address
}

func (c *Context) GetIssuer() [64]byte {
	return c.Issuer
}

func (c *Context) GetBalance() uint64 {
	return c.Balance
}

func (c *Context) GetSender() [64]byte {
	return c.From
}

func (c *Context) GetAmount() uint64 {
	return c.Amount
}

func (c *Context) GetTransactionData() []byte {
	return c.Data
}

func (c *Context) GetFee() uint64 {
	return c.Fee
}

func (c *Context) GetSig() [64]byte {
	return c.Sig
}
