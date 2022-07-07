package db

import (
	"fmt"
	"log"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

func GetNftClass(conn *pgxpool.Conn, pagination Pagination) (NftClassResponse, error) {
	sql := fmt.Sprintf(`
	SELECT class_id, name, description, symbol, uri, uri_hash,
		config, metadata, price,
		parent_type, parent_iscn_id_prefix, parent_account
	FROM nft_class
	WHERE	($1 = 0 OR id > $1)
		AND ($2 = 0 OR id < $2)
	ORDER BY id %s
	LIMIT $3
	`, pagination.Order)

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, pagination.After, pagination.Before, pagination.Limit)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	response := NftClassResponse{
		Classes: make([]NftClass, 0, pagination.Limit),
	}
	for rows.Next() {
		var c NftClass
		if err = rows.Scan(
			&c.Id, &c.Name, &c.Description, &c.Symbol, &c.URI, &c.URIHash,
			&c.Config, &c.Metadata, &c.Price,
			&c.Parent.Type, &c.Parent.IscnIdPrefix, &c.Parent.Account,
		); err != nil {
			panic(err)
		}
		log.Println(c)
		response.Classes = append(response.Classes, c)
	}
	return response, nil
}

func GetNftByIscn(conn *pgxpool.Conn, iscn string) (QueryNftByIscnResponse, error) {
	sql := `
	SELECT c.class_id, c.name, c.description, c.symbol, c.uri, c.uri_hash,
	c.config, c.metadata, c.price,
	c.parent_type, c.parent_iscn_id_prefix, c.parent_account,
	(
		SELECT array_agg(row_to_json((n.*)))
		FROM nft as n
		WHERE n.class_id = c.class_id
		GROUP BY n.class_id
	) as nfts
	FROM nft_class as c
	WHERE c.parent_iscn_id_prefix = $1
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, iscn)
	if err != nil {
		panic(err)
	}

	res := QueryNftByIscnResponse{
		IscnIdPrefix: iscn,
		Classes:      make([]QueryNftClassResponse, 0),
	}
	for rows.Next() {
		log.Println("hey")
		var c QueryNftClassResponse
		var nfts pgtype.JSONBArray
		if err = rows.Scan(
			&c.Id, &c.Name, &c.Description, &c.Symbol, &c.URI, &c.URIHash,
			&c.Config, &c.Metadata, &c.Price,
			&c.Parent.Type, &c.Parent.IscnIdPrefix, &c.Parent.Account, &nfts,
		); err != nil {
			panic(err)
		}
		if err = nfts.AssignTo(&c.Nfts); err != nil {
			panic(err)
		}
		c.Count = len(c.Nfts)
		res.Classes = append(res.Classes, c)
	}
	return res, nil
}
