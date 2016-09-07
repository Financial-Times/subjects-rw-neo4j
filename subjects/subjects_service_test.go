package subjects

import (
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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
	subjectsDriver := getSubjectsCypherDriver(t)
	err := subjectsDriver.Check()
	assert.NoError(t, err, "Unexpected error on connectivity check")
}

func TestPrefLabelIsCorrectlyWritten(t *testing.T) {
	subjectsDriver := getSubjectsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{UUIDS: []string{subjectUUID}}
	subjectToWrite := Subject{UUID: subjectUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	err := subjectsDriver.Write(subjectToWrite)
	assert.NoError(t, err, "ERROR happened during write time")

	storedSubject, found, err := subjectsDriver.Read(subjectUUID)
	assert.NoError(t, err, "ERROR happened during read time")
	assert.Equal(t, true, found)
	assert.NotEmpty(t, storedSubject)

	assert.Equal(t, prefLabel, storedSubject.(Subject).PrefLabel, "PrefLabel should be "+prefLabel)
	cleanUp(t, subjectUUID, subjectsDriver)
}

func TestPrefLabelSpecialCharactersAreHandledByCreate(t *testing.T) {
	subjectsDriver := getSubjectsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{}, UUIDS: []string{subjectUUID}}
	subjectToWrite := Subject{UUID: subjectUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(t, subjectsDriver.Write(subjectToWrite), "Failed to write subject")

	//add default types that will be automatically added by the writer
	subjectToWrite.Types = defaultTypes
	//check if subjectToWrite is the same with the one inside the DB
	readSubjectForUUIDAndCheckFieldsMatch(t, subjectsDriver, subjectUUID, subjectToWrite)
	cleanUp(t, subjectUUID, subjectsDriver)
}

func TestCreateCompleteSubjectWithPropsAndIdentifiers(t *testing.T) {
	subjectsDriver := getSubjectsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{subjectUUID}}
	subjectToWrite := Subject{UUID: subjectUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(t, subjectsDriver.Write(subjectToWrite), "Failed to write subject")

	//add default types that will be automatically added by the writer
	subjectToWrite.Types = defaultTypes
	//check if subjectToWrite is the same with the one inside the DB
	readSubjectForUUIDAndCheckFieldsMatch(t, subjectsDriver, subjectUUID, subjectToWrite)
	cleanUp(t, subjectUUID, subjectsDriver)
}

func TestUpdateWillRemovePropertiesAndIdentifiersNoLongerPresent(t *testing.T) {
	subjectsDriver := getSubjectsCypherDriver(t)

	allAlternativeIdentifiers := alternativeIdentifiers{TME: []string{}, UUIDS: []string{subjectUUID}}
	subjectToWrite := Subject{UUID: subjectUUID, PrefLabel: prefLabel, AlternativeIdentifiers: allAlternativeIdentifiers}

	assert.NoError(t, subjectsDriver.Write(subjectToWrite), "Failed to write subject")
	//add default types that will be automatically added by the writer
	subjectToWrite.Types = defaultTypes
	readSubjectForUUIDAndCheckFieldsMatch(t, subjectsDriver, subjectUUID, subjectToWrite)

	tmeAlternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{subjectUUID}}
	updatedSubject := Subject{UUID: subjectUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: tmeAlternativeIdentifiers}

	assert.NoError(t, subjectsDriver.Write(updatedSubject), "Failed to write updated subject")
	//add default types that will be automatically added by the writer
	updatedSubject.Types = defaultTypes
	readSubjectForUUIDAndCheckFieldsMatch(t, subjectsDriver, subjectUUID, updatedSubject)

	cleanUp(t, subjectUUID, subjectsDriver)
}

func TestDelete(t *testing.T) {
	subjectsDriver := getSubjectsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{subjectUUID}}
	subjectToDelete := Subject{UUID: subjectUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(t, subjectsDriver.Write(subjectToDelete), "Failed to write subject")

	found, err := subjectsDriver.Delete(subjectUUID)
	assert.True(t, found, "Didn't manage to delete subject for uuid %", subjectUUID)
	assert.NoError(t, err, "Error deleting subject for uuid %s", subjectUUID)

	p, found, err := subjectsDriver.Read(subjectUUID)

	assert.Equal(t, Subject{}, p, "Found subject %s who should have been deleted", p)
	assert.False(t, found, "Found subject for uuid %s who should have been deleted", subjectUUID)
	assert.NoError(t, err, "Error trying to find subject for uuid %s", subjectUUID)
}

func TestCount(t *testing.T) {
	subjectsDriver := getSubjectsCypherDriver(t)

	alternativeIds := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{subjectUUID}}
	subjectOneToCount := Subject{UUID: subjectUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIds}

	assert.NoError(t, subjectsDriver.Write(subjectOneToCount), "Failed to write subject")

	nr, err := subjectsDriver.Count()
	assert.Equal(t, 1, nr, "Should be 1 subjects in DB - count differs")
	assert.NoError(t, err, "An unexpected error occurred during count")

	newAlternativeIds := alternativeIdentifiers{TME: []string{newTmeID}, UUIDS: []string{newSubjectUUID}}
	subjectTwoToCount := Subject{UUID: newSubjectUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: newAlternativeIds}

	assert.NoError(t, subjectsDriver.Write(subjectTwoToCount), "Failed to write subject")

	nr, err = subjectsDriver.Count()
	assert.Equal(t, 2, nr, "Should be 2 subjects in DB - count differs")
	assert.NoError(t, err, "An unexpected error occurred during count")

	cleanUp(t, subjectUUID, subjectsDriver)
	cleanUp(t, newSubjectUUID, subjectsDriver)
}

func readSubjectForUUIDAndCheckFieldsMatch(t *testing.T, subjectsDriver service, uuid string, expectedSubject Subject) {

	storedSubject, found, err := subjectsDriver.Read(uuid)

	assert.NoError(t, err, "Error finding subject for uuid %s", uuid)
	assert.True(t, found, "Didn't find subject for uuid %s", uuid)
	assert.Equal(t, expectedSubject, storedSubject, "subjects should be the same")
}

func getSubjectsCypherDriver(t *testing.T) service {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	db, err := neoutils.Connect(url, conf)
	assert.NoError(t, err, "Failed to connect to Neo4j")
	return NewCypherSubjectsService(db)
}

func cleanUp(t *testing.T, uuid string, subjectsDriver service) {
	found, err := subjectsDriver.Delete(uuid)
	assert.True(t, found, "Didn't manage to delete subject for uuid %", uuid)
	assert.NoError(t, err, "Error deleting subject for uuid %s", uuid)
}
