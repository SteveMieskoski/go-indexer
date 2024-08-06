package consume

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
	AddAddressBalance(appDoc types.Address) (string, error)
	AddAddressDetail(appDoc types.Address) (string, error)
	AddContractAddress(appDoc types.Address) (string, error)
	AddAddressOnly(appDoc types.Address) (string, error)
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

// to check fields see this site: https://medium.com/@anajankow/fast-check-if-all-struct-fields-are-set-in-golang-bba1917213d2
func (a *addressRepository) Add(appDoc types.Address) (string, error) {

	_, err := a.client.Exec(context.Background(),
		`insert into addresses ("Address", "Nonce", "IsContract", "Balance", "LastSeen")
values ($1, $2, $3, $4, $5) ON CONFLICT ("Address") DO UPDATE SET "Nonce" = $2, "IsContract" = $3;`, appDoc.Address, appDoc.Nonce, appDoc.IsContract, appDoc.Balance, appDoc.LastSeen)
	if err != nil {
		utils.Logger.Errorln(err)
	}

	return strconv.Itoa(int(appDoc.Id)), nil
}

func (a *addressRepository) AddAddressOnly(appDoc types.Address) (string, error) {

	_, err := a.client.Exec(context.Background(),
		`insert into addresses ("Address", "LastSeen")
values ($1, $2) ON CONFLICT ("Address") DO UPDATE SET "LastSeen" = $2;`, appDoc.Address, appDoc.LastSeen)
	if err != nil {
		utils.Logger.Errorln(err)
	}

	return strconv.Itoa(int(appDoc.Id)), nil
}

func (a *addressRepository) AddContractAddress(appDoc types.Address) (string, error) {

	_, err := a.client.Exec(context.Background(),
		`insert into addresses ("Address", "IsContract", "LastSeen")
values ($1, $2, $3) ON CONFLICT ("Address") DO UPDATE SET "LastSeen" = $3;`, appDoc.Address, appDoc.IsContract, appDoc.LastSeen)
	if err != nil {
		utils.Logger.Errorln(err)
	}

	return strconv.Itoa(int(appDoc.Id)), nil
}

func (a *addressRepository) AddAddressDetail(appDoc types.Address) (string, error) {
	// maybe the nonce collection should get combined with the balance requests.
	_, err := a.client.Exec(context.Background(),
		`insert into addresses ("Address", "Nonce", "LastSeen", "IsContract")
values ($1, $2, $3, $4) ON CONFLICT ("Address") DO UPDATE SET "LastSeen" = $3, "Nonce" = $2  WHERE addresses."Nonce" <= $2;`, appDoc.Address, appDoc.Nonce, appDoc.LastSeen, appDoc.IsContract)
	if err != nil {
		utils.Logger.Errorln(err)
	}

	return strconv.Itoa(int(appDoc.Id)), nil
}

func (a *addressRepository) AddAddressBalance(appDoc types.Address) (string, error) {

	_, err := a.client.Exec(context.Background(),
		`insert into addresses ("Address",  "Balance", "LastSeen")
values ($1, $2, $3) ON CONFLICT ("Address") DO UPDATE SET "Balance" = $2, "LastSeen" = $3;`, appDoc.Address, appDoc.Balance, appDoc.LastSeen)
	if err != nil {
		utils.Logger.Errorln(err)
	}

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
			utils.Logger.Errorln(err)
			return err

		}
	}
	return nil
}

func (a *addressRepository) Delete(id string) (int64, error) {

	_, err := a.client.Exec(context.Background(),
		`delete from addresses where "Id" = $1;`, id)
	if err != nil {
		utils.Logger.Errorln(err)
	}

	return 0, nil
}
