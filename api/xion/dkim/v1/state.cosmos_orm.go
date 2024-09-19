// Code generated by protoc-gen-go-cosmos-orm. DO NOT EDIT.

package dkimv1

import (
	context "context"
	ormlist "cosmossdk.io/orm/model/ormlist"
	ormtable "cosmossdk.io/orm/model/ormtable"
	ormerrors "cosmossdk.io/orm/types/ormerrors"
)

type DkimPubKeyTable interface {
	Insert(ctx context.Context, dkimPubKey *DkimPubKey) error
	Update(ctx context.Context, dkimPubKey *DkimPubKey) error
	Save(ctx context.Context, dkimPubKey *DkimPubKey) error
	Delete(ctx context.Context, dkimPubKey *DkimPubKey) error
	Has(ctx context.Context, domain string, selector string) (found bool, err error)
	// Get returns nil and an error which responds true to ormerrors.IsNotFound() if the record was not found.
	Get(ctx context.Context, domain string, selector string) (*DkimPubKey, error)
	List(ctx context.Context, prefixKey DkimPubKeyIndexKey, opts ...ormlist.Option) (DkimPubKeyIterator, error)
	ListRange(ctx context.Context, from, to DkimPubKeyIndexKey, opts ...ormlist.Option) (DkimPubKeyIterator, error)
	DeleteBy(ctx context.Context, prefixKey DkimPubKeyIndexKey) error
	DeleteRange(ctx context.Context, from, to DkimPubKeyIndexKey) error

	doNotImplement()
}

type DkimPubKeyIterator struct {
	ormtable.Iterator
}

func (i DkimPubKeyIterator) Value() (*DkimPubKey, error) {
	var dkimPubKey DkimPubKey
	err := i.UnmarshalMessage(&dkimPubKey)
	return &dkimPubKey, err
}

type DkimPubKeyIndexKey interface {
	id() uint32
	values() []interface{}
	dkimPubKeyIndexKey()
}

// primary key starting index..
type DkimPubKeyPrimaryKey = DkimPubKeyDomainSelectorIndexKey

type DkimPubKeyDomainSelectorIndexKey struct {
	vs []interface{}
}

func (x DkimPubKeyDomainSelectorIndexKey) id() uint32            { return 0 }
func (x DkimPubKeyDomainSelectorIndexKey) values() []interface{} { return x.vs }
func (x DkimPubKeyDomainSelectorIndexKey) dkimPubKeyIndexKey()   {}

func (this DkimPubKeyDomainSelectorIndexKey) WithDomain(domain string) DkimPubKeyDomainSelectorIndexKey {
	this.vs = []interface{}{domain}
	return this
}

func (this DkimPubKeyDomainSelectorIndexKey) WithDomainSelector(domain string, selector string) DkimPubKeyDomainSelectorIndexKey {
	this.vs = []interface{}{domain, selector}
	return this
}

type dkimPubKeyTable struct {
	table ormtable.Table
}

func (this dkimPubKeyTable) Insert(ctx context.Context, dkimPubKey *DkimPubKey) error {
	return this.table.Insert(ctx, dkimPubKey)
}

func (this dkimPubKeyTable) Update(ctx context.Context, dkimPubKey *DkimPubKey) error {
	return this.table.Update(ctx, dkimPubKey)
}

func (this dkimPubKeyTable) Save(ctx context.Context, dkimPubKey *DkimPubKey) error {
	return this.table.Save(ctx, dkimPubKey)
}

func (this dkimPubKeyTable) Delete(ctx context.Context, dkimPubKey *DkimPubKey) error {
	return this.table.Delete(ctx, dkimPubKey)
}

func (this dkimPubKeyTable) Has(ctx context.Context, domain string, selector string) (found bool, err error) {
	return this.table.PrimaryKey().Has(ctx, domain, selector)
}

func (this dkimPubKeyTable) Get(ctx context.Context, domain string, selector string) (*DkimPubKey, error) {
	var dkimPubKey DkimPubKey
	found, err := this.table.PrimaryKey().Get(ctx, &dkimPubKey, domain, selector)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ormerrors.NotFound
	}
	return &dkimPubKey, nil
}

func (this dkimPubKeyTable) List(ctx context.Context, prefixKey DkimPubKeyIndexKey, opts ...ormlist.Option) (DkimPubKeyIterator, error) {
	it, err := this.table.GetIndexByID(prefixKey.id()).List(ctx, prefixKey.values(), opts...)
	return DkimPubKeyIterator{it}, err
}

func (this dkimPubKeyTable) ListRange(ctx context.Context, from, to DkimPubKeyIndexKey, opts ...ormlist.Option) (DkimPubKeyIterator, error) {
	it, err := this.table.GetIndexByID(from.id()).ListRange(ctx, from.values(), to.values(), opts...)
	return DkimPubKeyIterator{it}, err
}

func (this dkimPubKeyTable) DeleteBy(ctx context.Context, prefixKey DkimPubKeyIndexKey) error {
	return this.table.GetIndexByID(prefixKey.id()).DeleteBy(ctx, prefixKey.values()...)
}

func (this dkimPubKeyTable) DeleteRange(ctx context.Context, from, to DkimPubKeyIndexKey) error {
	return this.table.GetIndexByID(from.id()).DeleteRange(ctx, from.values(), to.values())
}

func (this dkimPubKeyTable) doNotImplement() {}

var _ DkimPubKeyTable = dkimPubKeyTable{}

func NewDkimPubKeyTable(db ormtable.Schema) (DkimPubKeyTable, error) {
	table := db.GetTable(&DkimPubKey{})
	if table == nil {
		return nil, ormerrors.TableNotFound.Wrap(string((&DkimPubKey{}).ProtoReflect().Descriptor().FullName()))
	}
	return dkimPubKeyTable{table}, nil
}

type StateStore interface {
	DkimPubKeyTable() DkimPubKeyTable

	doNotImplement()
}

type stateStore struct {
	dkimPubKey DkimPubKeyTable
}

func (x stateStore) DkimPubKeyTable() DkimPubKeyTable {
	return x.dkimPubKey
}

func (stateStore) doNotImplement() {}

var _ StateStore = stateStore{}

func NewStateStore(db ormtable.Schema) (StateStore, error) {
	dkimPubKeyTable, err := NewDkimPubKeyTable(db)
	if err != nil {
		return nil, err
	}

	return stateStore{
		dkimPubKeyTable,
	}, nil
}
