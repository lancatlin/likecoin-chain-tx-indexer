package rest

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleISCN(c *gin.Context) {
	q := c.Request.URL.Query()
	if q.Get("q") != "" {
		handleISCNSearch(c)
		return
	}

	var iscn db.ISCN
	hasQuery := false

	for k, v := range q {
		switch k {
		case "iscn_id":
			if len(v) > 0 {
				hasQuery = true
				iscn.Iscn = v[0]
			}
		case "owner":
			if len(v) > 0 {
				hasQuery = true
				iscn.Owner = v[0]
			}
		case "fingerprint", "fingerprints":
			hasQuery = true
			iscn.Fingerprints = v
		case "keywords":
			hasQuery = true
			iscn.Keywords = v
		case "stakeholders.entity.id", "stakeholders.id":
			if len(v) > 0 {
				hasQuery = true
				iscn.Stakeholders = []byte(fmt.Sprintf(`[{"id": "%s"}]`, v[0]))
			}
		case "stakeholders.entity.name", "stakeholders.name":
			if len(v) > 0 {
				hasQuery = true
				iscn.Stakeholders = []byte(fmt.Sprintf(`[{"name": "%s"}]`, v[0]))
			}
		}
	}

	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	pool := getPool(c)
	var res db.ISCNResponse
	if hasQuery {
		res, err = db.QueryISCN(pool, iscn, p)
	} else {
		res, err = db.QueryISCNList(pool, p)
	}
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleISCNSearch(c *gin.Context) {
	q := c.Request.URL.Query()
	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	term := q.Get("q")
	if term == "" {
		c.AbortWithStatusJSON(404, gin.H{"error": "parameter 'q' is required"})
		return
	}
	pool := getPool(c)
	res, err := db.QueryISCNAll(pool, term, p)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
