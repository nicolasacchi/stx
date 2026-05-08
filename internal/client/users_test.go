package client

import (
	"context"
	"testing"
)

func TestListUsers(t *testing.T) {
	rs := newRecordingServer(t, 200, `[]`)
	defer rs.close()
	c := rs.client("k", "")
	c.ListUsers(context.Background())
	rs.assertRequest(t, "GET", "/api/v2/users")
}

func TestGetUser(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.GetUser(context.Background(), "user-id")
	rs.assertRequest(t, "GET", "/api/v2/users/user-id")
}

func TestCreateUser(t *testing.T) {
	rs := newRecordingServer(t, 201, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.CreateUser(context.Background(), map[string]string{"username": "u", "password": "p"})
	rs.assertRequest(t, "POST", "/api/v2/users")
	got := rs.bodyJSON()
	if got["username"] != "u" || got["password"] != "p" {
		t.Errorf("body: %v", got)
	}
}

func TestAttachUser(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.AttachUser(context.Background(), map[string]any{"email": "x@y.z", "hasNoProducts": true})
	rs.assertRequest(t, "POST", "/api/v2/users/attach-user")
}

func TestDetachUser(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.DetachUser(context.Background(), "user-id")
	rs.assertRequest(t, "POST", "/api/v2/users/user-id/detach-user")
	if len(rs.body) != 0 {
		t.Errorf("detach has no body, got %q", rs.body)
	}
}

func TestExportUsersCSV(t *testing.T) {
	rs := newRecordingServer(t, 200, "id,name\n1,foo\n")
	defer rs.close()
	c := rs.client("k", "")
	data, err := c.ExportUsersCSV(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	rs.assertRequest(t, "GET", "/api/v2/users/export-csv")
	if string(data) != "id,name\n1,foo\n" {
		t.Errorf("got CSV %q", data)
	}
}
