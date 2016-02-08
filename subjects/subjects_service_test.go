package subjects

import (
	"fmt"
	"os"
	"testing"

	"github.com/Financial-Times/base-ft-rw-app-go"
	"github.com/Financial-Times/neo-utils-go"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

var subjectsDriver baseftrwapp.Service

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"

	subjectsDriver = getSubjectsCypherDriver(t)

	subjectToDelete := Subject{UUID: uuid, CanonicalName: "Test", TmeIdentifier: "TME_ID"}

	assert.NoError(subjectsDriver.Write(subjectToDelete), "Failed to write subject")

	found, err := subjectsDriver.Delete(uuid)
	assert.True(found, "Didn't manage to delete subject for uuid %", uuid)
	assert.NoError(err, "Error deleting subject for uuid %s", uuid)

	p, found, err := subjectsDriver.Read(uuid)

	assert.Equal(Subject{}, p, "Found subject %s who should have been deleted", p)
	assert.False(found, "Found subject for uuid %s who should have been deleted", uuid)
	assert.NoError(err, "Error trying to find subject for uuid %s", uuid)
}

func TestCreateAllValuesPresent(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	subjectsDriver = getSubjectsCypherDriver(t)

	subjectToWrite := Subject{UUID: uuid, CanonicalName: "Test", TmeIdentifier: "TME_ID"}

	assert.NoError(subjectsDriver.Write(subjectToWrite), "Failed to write subject")

	readSubjectForUUIDAndCheckFieldsMatch(t, uuid, subjectToWrite)

	cleanUp(t, uuid)
}

func TestCreateHandlesSpecialCharacters(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	subjectsDriver = getSubjectsCypherDriver(t)

	subjectToWrite := Subject{UUID: uuid, CanonicalName: "Test 'special chars", TmeIdentifier: "TME_ID"}

	assert.NoError(subjectsDriver.Write(subjectToWrite), "Failed to write subject")

	readSubjectForUUIDAndCheckFieldsMatch(t, uuid, subjectToWrite)

	cleanUp(t, uuid)
}

func TestCreateNotAllValuesPresent(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	subjectsDriver = getSubjectsCypherDriver(t)

	subjectToWrite := Subject{UUID: uuid, CanonicalName: "Test"}

	assert.NoError(subjectsDriver.Write(subjectToWrite), "Failed to write subject")

	readSubjectForUUIDAndCheckFieldsMatch(t, uuid, subjectToWrite)

	cleanUp(t, uuid)
}

func TestUpdateWillRemovePropertiesNoLongerPresent(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	subjectsDriver = getSubjectsCypherDriver(t)

	subjectToWrite := Subject{UUID: uuid, CanonicalName: "Test", TmeIdentifier: "TME_ID"}

	assert.NoError(subjectsDriver.Write(subjectToWrite), "Failed to write subject")
	readSubjectForUUIDAndCheckFieldsMatch(t, uuid, subjectToWrite)

	updatedSubject := Subject{UUID: uuid, CanonicalName: "Test", TmeIdentifier: "TME_ID"}

	assert.NoError(subjectsDriver.Write(updatedSubject), "Failed to write updated subject")
	readSubjectForUUIDAndCheckFieldsMatch(t, uuid, updatedSubject)

	cleanUp(t, uuid)
}

func TestConnectivityCheck(t *testing.T) {
	assert := assert.New(t)
	subjectsDriver = getSubjectsCypherDriver(t)
	err := subjectsDriver.Check()
	assert.NoError(err, "Unexpected error on connectivity check")
}

func getSubjectsCypherDriver(t *testing.T) service {
	assert := assert.New(t)
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	db, err := neoism.Connect(url)
	assert.NoError(err, "Failed to connect to Neo4j")
	return NewCypherSubjectsService(neoutils.StringerDb{db}, db)
}

func readSubjectForUUIDAndCheckFieldsMatch(t *testing.T, uuid string, expectedSubject Subject) {
	assert := assert.New(t)
	storedSubject, found, err := subjectsDriver.Read(uuid)

	assert.NoError(err, "Error finding subject for uuid %s", uuid)
	assert.True(found, "Didn't find subject for uuid %s", uuid)
	assert.Equal(expectedSubject, storedSubject, "subjects should be the same")
}

func TestWritePrefLabelIsAlsoWrittenAndIsEqualToName(t *testing.T) {
	assert := assert.New(t)
	subjectsDriver := getSubjectsCypherDriver(t)
	uuid := "12345"
	subjectToWrite := Subject{UUID: uuid, CanonicalName: "Test", TmeIdentifier: "TME_ID"}

	storedSubject := subjectsDriver.Write(subjectToWrite)

	fmt.Printf("", storedSubject)

	result := []struct {
		PrefLabel string `json:"t.prefLabel"`
	}{}

	getPrefLabelQuery := &neoism.CypherQuery{
		Statement: `
				MATCH (t:Subject {uuid:"12345"}) RETURN t.prefLabel
				`,
		Result: &result,
	}

	err := subjectsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{getPrefLabelQuery})
	assert.NoError(err)
	assert.Equal("Test", result[0].PrefLabel, "PrefLabel should be 'Test")
	cleanUp(t, uuid)
}

func cleanUp(t *testing.T, uuid string) {
	assert := assert.New(t)
	found, err := subjectsDriver.Delete(uuid)
	assert.True(found, "Didn't manage to delete subject for uuid %", uuid)
	assert.NoError(err, "Error deleting subject for uuid %s", uuid)
}
