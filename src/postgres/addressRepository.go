package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"src/types"
	"src/utils"
	"strconv"
)

type AddressRepository interface {
	Add(appDoc types.Address) (string, error)
	//List(count int, ctx context.Context) ([]*types.Address, error)
	//GetById(oId string, ctx context.Context) (*types.Address, error)
	Update(appDoc types.Address) error
	Delete(oId string) (int64, error)
}

type addressRepository struct {
	client       *pgxpool.Pool
	indicesExist bool
}

func NewAddressRepository(client *PostgresDB) AddressRepository {
	return &addressRepository{client: client.client, indicesExist: false}
}

func (a *addressRepository) Add(appDoc types.Address) (string, error) {

	_, err := a.client.Exec(context.Background(),
		`insert into addresses ("Address", "Nonce", "IsContract", "Balance", "LastSeen")
values ($1, $2, $3, $4);`, appDoc.Address, appDoc.Nonce, appDoc.IsContract, appDoc.Balance, appDoc.LastSeen)
	if err != nil {
		utils.Logger.Errorln(err)
	}
	//
	//
	//if result.Error != nil {
	//	return "", result.Error
	//}

	return strconv.Itoa(int(appDoc.Id)), nil
}

func (a *addressRepository) Update(appDoc types.Address) error {
	var updateString = `
		update addresses 
		set "Nonce" = $2, "LastSeen" = $3
		where "Address" = $1 AND "LastSeen" < $3;`

	_, err := a.client.Exec(context.Background(), updateString,
		appDoc.Address, appDoc.Nonce, appDoc.LastSeen)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// there were no rows, but otherwise no error occurred
		} else {
			//panic(err)
			utils.Logger.Errorln(err)
			return err

		}
	}
	return nil
}

// func (a *addressRepository) List(count int, ctx context.Context) ([]*types.Address, error) {
//
//		var addresses []*types.Address
//		result := a.client.Find(&addresses)
//
//		if result.Error != nil {
//			return nil, result.Error
//		}
//
//		return addresses, nil
//	}
//
// func (a *addressRepository) GetById(id string, ctx context.Context) (*types.Address, error) {
//
//		var address *types.Address
//
//		a.client.First(&address, id)
//
//		return address, nil
//	}
func (a *addressRepository) Delete(id string) (int64, error) {

	_, err := a.client.Exec(context.Background(),
		`delete from addresses where "Id" = $1;`, id)
	if err != nil {
		utils.Logger.Errorln(err)
	}

	return 0, nil
}
