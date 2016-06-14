package subjects

import (
	"os"
	"testing"

	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

const (
	subjectUUID          = "12345"
	newSubjectUUID       = "123456"
	tmeID                = "TME_ID"
	newTmeID             = "NEW_TME_ID"
	prefLabel            = "Test"
	specialCharPrefLabel = "Test 'special chars"
)

var defaultTypes = []string{"Thing", "Concept", "Classification", "Subject"}

func TestConnectivityCheck(t *testing.T) {
	assert := assert.New(t)
	subjectsDriver := getSubjectsCypherDriver(t)
	err := subjectsDriver.Check()
	assert.NoError(err, "Unexpected error on connectivity check")
}

func TestPrefLabelIsCorrectlyWritten(t *testing.T) {
	assert := assert.New(t)
	subjectsDriver := getSubjectsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{UUIDS: []string{subjectUUID}}
	subjectToWrite := Subject{UUID: subjectUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	err := subjectsDriver.Write(subjectToWrite)
	assert.NoError(err, "ERROR happened during write time")

	storedSubject, found, err := subjectsDriver.Read(subjectUUID)
	assert.NoError(err, "ERROR happened during read time")
	assert.Equal(true, found)
	assert.NotEmpty(storedSubject)

	assert.Equal(prefLabel, storedSubject.(Subject).PrefLabel, "PrefLabel should be "+prefLabel)
	cleanUp(assert, subjectUUID, subjectsDriver)
}

func TestPrefLabelSpecialCharactersAreHandledByCreate(t *testing.T) {
	assert := assert.New(t)
	subjectsDriver := getSubjectsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{}, UUIDS: []string{subjectUUID}}
	subjectToWrite := Subject{UUID: subjectUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(subjectsDriver.Write(subjectToWrite), "Failed to write subject")

	//add default types that will be automatically added by the writer
	subjectToWrite.Types = defaultTypes
	//check if subjectToWrite is the same with the one inside the DB
	readSubjectForUUIDAndCheckFieldsMatch(assert, subjectsDriver, subjectUUID, subjectToWrite)
	cleanUp(assert, subjectUUID, subjectsDriver)
}

func TestCreateCompleteSubjectWithPropsAndIdentifiers(t *testing.T) {
	assert := assert.New(t)
	subjectsDriver := getSubjectsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{subjectUUID}}
	subjectToWrite := Subject{UUID: subjectUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(subjectsDriver.Write(subjectToWrite), "Failed to write subject")

	//add default types that will be automatically added by the writer
	subjectToWrite.Types = defaultTypes
	//check if subjectToWrite is the same with the one inside the DB
	readSubjectForUUIDAndCheckFieldsMatch(assert, subjectsDriver, subjectUUID, subjectToWrite)
	cleanUp(assert, subjectUUID, subjectsDriver)
}

func TestUpdateWillRemovePropertiesAndIdentifiersNoLongerPresent(t *testing.T) {
	assert := assert.New(t)
	subjectsDriver := getSubjectsCypherDriver(t)

	allAlternativeIdentifiers := alternativeIdentifiers{TME: []string{}, UUIDS: []string{subjectUUID}}
	subjectToWrite := Subject{UUID: subjectUUID, PrefLabel: prefLabel, AlternativeIdentifiers: allAlternativeIdentifiers}

	assert.NoError(subjectsDriver.Write(subjectToWrite), "Failed to write subject")
	//add default types that will be automatically added by the writer
	subjectToWrite.Types = defaultTypes
	readSubjectForUUIDAndCheckFieldsMatch(assert, subjectsDriver, subjectUUID, subjectToWrite)

	tmeAlternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{subjectUUID}}
	updatedSubject := Subject{UUID: subjectUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: tmeAlternativeIdentifiers}

	assert.NoError(subjectsDriver.Write(updatedSubject), "Failed to write updated subject")
	//add default types that will be automatically added by the writer
	updatedSubject.Types = defaultTypes
	readSubjectForUUIDAndCheckFieldsMatch(assert, subjectsDriver, subjectUUID, updatedSubject)

	cleanUp(assert, subjectUUID, subjectsDriver)
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	subjectsDriver := getSubjectsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{subjectUUID}}
	subjectToDelete := Subject{UUID: subjectUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(subjectsDriver.Write(subjectToDelete), "Failed to write subject")

	found, err := subjectsDriver.Delete(subjectUUID)
	assert.True(found, "Didn't manage to delete subject for uuid %", subjectUUID)
	assert.NoError(err, "Error deleting subject for uuid %s", subjectUUID)

	p, found, err := subjectsDriver.Read(subjectUUID)

	assert.Equal(Subject{}, p, "Found subject %s who should have been deleted", p)
	assert.False(found, "Found subject for uuid %s who should have been deleted", subjectUUID)
	assert.NoError(err, "Error trying to find subject for uuid %s", subjectUUID)
}

func TestCount(t *testing.T) {
	assert := assert.New(t)
	subjectsDriver := getSubjectsCypherDriver(t)

	alternativeIds := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{subjectUUID}}
	subjectOneToCount := Subject{UUID: subjectUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIds}

	assert.NoError(subjectsDriver.Write(subjectOneToCount), "Failed to write subject")

	nr, err := subjectsDriver.Count()
	assert.Equal(1, nr, "Should be 1 subjects in DB - count differs")
	assert.NoError(err, "An unexpected error occurred during count")

	newAlternativeIds := alternativeIdentifiers{TME: []string{newTmeID}, UUIDS: []string{newSubjectUUID}}
	subjectTwoToCount := Subject{UUID: newSubjectUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: newAlternativeIds}

	assert.NoError(subjectsDriver.Write(subjectTwoToCount), "Failed to write subject")

	nr, err = subjectsDriver.Count()
	assert.Equal(2, nr, "Should be 2 subjects in DB - count differs")
	assert.NoError(err, "An unexpected error occurred during count")

	cleanUp(assert, subjectUUID, subjectsDriver)
	cleanUp(assert, newSubjectUUID, subjectsDriver)
}

func readSubjectForUUIDAndCheckFieldsMatch(assert *assert.Assertions, subjectsDriver service, uuid string, expectedSubject Subject) {

	storedSubject, found, err := subjectsDriver.Read(uuid)

	assert.NoError(err, "Error finding subject for uuid %s", uuid)
	assert.True(found, "Didn't find subject for uuid %s", uuid)
	assert.Equal(expectedSubject, storedSubject, "subjects should be the same")
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

func cleanUp(assert *assert.Assertions, uuid string, subjectsDriver service) {
	found, err := subjectsDriver.Delete(uuid)
	assert.True(found, "Didn't manage to delete subject for uuid %", uuid)
	assert.NoError(err, "Error deleting subject for uuid %s", uuid)
}
