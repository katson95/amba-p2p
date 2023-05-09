package cassandra

import (
	"testing"
)

func TestCassandraDB(t *testing.T) {
	cassandraInstaller := New()

	err := cassandraInstaller.Update()
	if err != nil {
		return
	}

	err = cassandraInstaller.PreReqJDK()
	if err != nil {
		return
	}

	err = cassandraInstaller.AddRepository()
	if err != nil {
		return
	}

	err = cassandraInstaller.AddKey()
	if err != nil {
		return
	}

	err = cassandraInstaller.Install()
	if err != nil {
		return
	}

	err = cassandraInstaller.Status()
	if err != nil {
		return
	}

}
