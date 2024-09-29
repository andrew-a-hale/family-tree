package database

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Db struct {
	Driver neo4j.DriverWithContext
	ctx    context.Context
}

func (db *Db) Context() context.Context {
	return db.ctx
}

func InitNeo4jDriver() *Db {
	if os.Getenv("NEO4J_URI") == "" {
		log.Fatal("Missing ENV VAR: NEO4J_URI")
	}

	if os.Getenv("NEO4J_USER") == "" {
		log.Fatal("Missing ENV VAR: NEO4J_USER")
	}

	if os.Getenv("NEO4J_PASSWORD") == "" {
		log.Fatal("Missing ENV VAR: NEO4J_PASSWORD")
	}

	ctx := context.Background()
	neo4jUri := os.Getenv("NEO4J_URI")
	neo4jUser := os.Getenv("NEO4J_USER")
	neo4jPassword := os.Getenv("NEO4J_PASSWORD")
	driver, err := neo4j.NewDriverWithContext(
		neo4jUri,
		neo4j.BasicAuth(neo4jUser, neo4jPassword, ""),
	)
	if err != nil {
		log.Fatal("Error: Failed to connect to Neo4j Database")
	}

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		log.Fatalf("Error: Unhealth Connection: %v\n", err)
	}

	return &Db{driver, ctx}
}

type Person struct {
	Node neo4j.Node
}

func (db *Db) GetPerson(r *http.Request) ([]Person, error) {
	ctx := context.Background()
	session := db.Driver.NewSession(r.Context(), neo4j.SessionConfig{})
	defer session.Close(r.Context())

	cypher := "match (p:Person) return p limit 2"
	res, err := neo4j.ExecuteRead(
		ctx,
		session,
		func(tx neo4j.ManagedTransaction) ([]Person, error) {
			res, err := tx.Run(ctx, cypher, nil)
			if err != nil {
				return nil, err
			}
			return neo4j.CollectTWithContext(
				ctx,
				res,
				func(record *neo4j.Record) (Person, error) {
					person, isNil, err := neo4j.GetRecordValue[neo4j.Node](record, "p")
					if isNil {
						fmt.Println("person value is nil")
					}
					if err != nil {
						return Person{}, err
					}
					return Person{person}, nil
				},
			)
		},
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}
