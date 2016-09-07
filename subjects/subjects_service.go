package subjects

import (
	"encoding/json"
	"fmt"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
)

type service struct {
	conn neoutils.NeoConnection
}

// NewCypherSubjectsService provides functions for create, update, delete operations on subjects in Neo4j,
// plus other utility functions needed for a service
func NewCypherSubjectsService(conn neoutils.NeoConnection) service {
	return service{conn}
}

func (s service) Initialise() error {
	return s.conn.EnsureConstraints(map[string]string{
		"Thing":          "uuid",
		"Concept":        "uuid",
		"Classification": "uuid",
		"Subject":        "uuid",
		"TMEIdentifier":  "value",
		"UPPIdentifier":  "value"})
}

func (s service) Read(uuid string) (interface{}, bool, error) {
	results := []Subject{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:Subject {uuid:{uuid}})
OPTIONAL MATCH (upp:UPPIdentifier)-[:IDENTIFIES]->(n)
OPTIONAL MATCH (tme:TMEIdentifier)-[:IDENTIFIES]->(n)
return distinct n.uuid as uuid, n.prefLabel as prefLabel, labels(n) as types, {uuids:collect(distinct upp.value), TME:collect(distinct tme.value)} as alternativeIdentifiers`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
		Result: &results,
	}

	err := s.conn.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return Subject{}, false, err
	}

	if len(results) == 0 {
		return Subject{}, false, nil
	}

	return results[0], true, nil
}

func (s service) Write(thing interface{}) error {

	subject := thing.(Subject)

	//cleanUP all the previous IDENTIFIERS referring to that uuid
	deletePreviousIdentifiersQuery := &neoism.CypherQuery{
		Statement: `MATCH (t:Thing {uuid:{uuid}})
		OPTIONAL MATCH (t)<-[iden:IDENTIFIES]-(i)
		DELETE iden, i`,
		Parameters: map[string]interface{}{
			"uuid": subject.UUID,
		},
	}

	//create-update node for Subject
	createSubjectQuery := &neoism.CypherQuery{
		Statement: `MERGE (n:Thing {uuid: {uuid}})
					set n={allprops}
					set n :Concept
					set n :Classification
					set n :Subject
		`,
		Parameters: map[string]interface{}{
			"uuid": subject.UUID,
			"allprops": map[string]interface{}{
				"uuid":      subject.UUID,
				"prefLabel": subject.PrefLabel,
			},
		},
	}

	queryBatch := []*neoism.CypherQuery{deletePreviousIdentifiersQuery, createSubjectQuery}

	//ADD all the IDENTIFIER nodes and IDENTIFIES relationships
	for _, alternativeUUID := range subject.AlternativeIdentifiers.TME {
		alternativeIdentifierQuery := createNewIdentifierQuery(subject.UUID, tmeIdentifierLabel, alternativeUUID)
		queryBatch = append(queryBatch, alternativeIdentifierQuery)
	}

	for _, alternativeUUID := range subject.AlternativeIdentifiers.UUIDS {
		alternativeIdentifierQuery := createNewIdentifierQuery(subject.UUID, uppIdentifierLabel, alternativeUUID)
		queryBatch = append(queryBatch, alternativeIdentifierQuery)
	}
	return s.conn.CypherBatch(queryBatch)
}

func createNewIdentifierQuery(uuid string, identifierLabel string, identifierValue string) *neoism.CypherQuery {
	statementTemplate := fmt.Sprintf(`MERGE (t:Thing {uuid:{uuid}})
					CREATE (i:Identifier {value:{value}})
					MERGE (t)<-[:IDENTIFIES]-(i)
					set i : %s `, identifierLabel)
	query := &neoism.CypherQuery{
		Statement: statementTemplate,
		Parameters: map[string]interface{}{
			"uuid":  uuid,
			"value": identifierValue,
		},
	}
	return query
}

func (s service) Delete(uuid string) (bool, error) {
	clearNode := &neoism.CypherQuery{
		Statement: `
			MATCH (s:Thing {uuid: {uuid}})
			OPTIONAL MATCH (s)<-[iden:IDENTIFIES]-(i:Identifier)
			REMOVE s:Concept
			REMOVE s:Classification
			REMOVE s:Subject
			DELETE iden, i
			SET s={uuid:{uuid}}
		`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
			"props": map[string]interface{}{
				"uuid": uuid,
			},
		},
		IncludeStats: true,
	}

	removeNodeIfUnused := &neoism.CypherQuery{
		Statement: `
			MATCH (s:Thing {uuid: {uuid}})
			OPTIONAL MATCH (s)-[a]-(x)
			WITH s, count(a) AS relCount
			WHERE relCount = 0
			DELETE s
		`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
	}

	err := s.conn.CypherBatch([]*neoism.CypherQuery{clearNode, removeNodeIfUnused})

	s1, err := clearNode.Stats()
	if err != nil {
		return false, err
	}

	var deleted bool
	if s1.ContainsUpdates && s1.LabelsRemoved > 0 {
		deleted = true
	}

	return deleted, err
}

func (s service) DecodeJSON(dec *json.Decoder) (interface{}, string, error) {
	sub := Subject{}
	err := dec.Decode(&sub)
	return sub, sub.UUID, err
}

func (s service) Check() error {
	return neoutils.Check(s.conn)
}

func (s service) Count() (int, error) {

	results := []struct {
		Count int `json:"c"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:Subject) return count(n) as c`,
		Result:    &results,
	}

	err := s.conn.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return 0, err
	}

	return results[0].Count, nil
}
