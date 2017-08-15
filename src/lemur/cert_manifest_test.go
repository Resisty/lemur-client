package main

import (
    "testing"
)

func TestNewCertManifest(t *testing.T) {
    // This should create a new *certManifest with no errors
    man := NewCertManifest("test_authority",
                           "common.name",
                           "email@example.com",
                           "1970-01-01",
                           "1970-01-02",
                           "TestOrg")
    if man.Authority["name"] != "test_authority" ||
       man.CommonName != "common.name" ||
       man.Email != "email@example.com" ||
       man.StartDate != "1970-01-01" ||
       man.EndDate != "1970-01-02" {
        t.Errorf("NewCertManifest should result in a *certManifest with populated fields!")
    }
    desc := man.Description
    man.makeDigest()
    if desc != man.Description {
        t.Errorf("Calls to .makeDigest() should not change the description!")
    }
}
