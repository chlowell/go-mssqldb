package mssql

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"net"
	"net/url"
	"os"
	"runtime"
	"testing"
)

type MockTransport struct {
	bytes.Buffer
}

func (t *MockTransport) Close() error {
	return nil
}

func TestSendLogin(t *testing.T) {
	memBuf := new(MockTransport)
	buf := newTdsBuffer(1024, memBuf)
	login := login{
		TDSVersion:     verTDS73,
		PacketSize:     0x1000,
		ClientProgVer:  0x01060100,
		ClientPID:      100,
		ClientTimeZone: -4 * 60,
		ClientID:       [6]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab},
		OptionFlags1:   0xe0,
		OptionFlags3:   8,
		HostName:       "subdev1",
		UserName:       "test",
		Password:       "testpwd",
		AppName:        "appname",
		ServerName:     "servername",
		CtlIntName:     "library",
		Language:       "en",
		Database:       "database",
		ClientLCID:     0x204,
		AtchDBFile:     "filepath",
	}
	err := sendLogin(buf, login)
	if err != nil {
		t.Error("sendLogin should succeed")
	}
	ref := []byte{
		16, 1, 0, 222, 0, 0, 1, 0, 198 + 16, 0, 0, 0, 3, 0, 10, 115, 0, 16, 0, 0, 0, 1,
		6, 1, 100, 0, 0, 0, 0, 0, 0, 0, 224, 0, 0, 8, 16, 255, 255, 255, 4, 2, 0,
		0, 94, 0, 7, 0, 108, 0, 4, 0, 116, 0, 7, 0, 130, 0, 7, 0, 144, 0, 10, 0, 0,
		0, 0, 0, 164, 0, 7, 0, 178, 0, 2, 0, 182, 0, 8, 0, 18, 52, 86, 120, 144, 171,
		198, 0, 0, 0, 198, 0, 8, 0, 214, 0, 0, 0, 0, 0, 0, 0, 115, 0, 117, 0, 98,
		0, 100, 0, 101, 0, 118, 0, 49, 0, 116, 0, 101, 0, 115, 0, 116, 0, 226, 165,
		243, 165, 146, 165, 226, 165, 162, 165, 210, 165, 227, 165, 97, 0, 112,
		0, 112, 0, 110, 0, 97, 0, 109, 0, 101, 0, 115, 0, 101, 0, 114, 0, 118, 0,
		101, 0, 114, 0, 110, 0, 97, 0, 109, 0, 101, 0, 108, 0, 105, 0, 98, 0, 114,
		0, 97, 0, 114, 0, 121, 0, 101, 0, 110, 0, 100, 0, 97, 0, 116, 0, 97, 0, 98,
		0, 97, 0, 115, 0, 101, 0, 102, 0, 105, 0, 108, 0, 101, 0, 112, 0, 97, 0,
		116, 0, 104, 0}
	out := memBuf.Bytes()
	if !bytes.Equal(ref, out) {
		t.Log("Expected:")
		t.Log(hex.Dump(ref))
		t.Log("Returned:")
		t.Log(hex.Dump(out))
		t.Fatal("input output don't match")
	}
}

func TestSendLoginWithFeatureExt(t *testing.T) {
	memBuf := new(MockTransport)
	buf := newTdsBuffer(1024, memBuf)
	login := login{
		TDSVersion:     verTDS74,
		PacketSize:     0x1000,
		ClientProgVer:  0x01060100,
		ClientPID:      100,
		ClientTimeZone: -4 * 60,
		ClientID:       [6]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab},
		OptionFlags1:   0xe0,
		OptionFlags3:   8,
		HostName:       "subdev1",
		AppName:        "appname",
		ServerName:     "servername",
		CtlIntName:     "library",
		Language:       "en",
		Database:       "database",
		ClientLCID:     0x204,
	}
	login.FeatureExt.Add(&featureExtFedAuthSTS{
		FedAuthToken: "fedauthtoken",
	})
	err := sendLogin(buf, login)
	if err != nil {
		t.Error("sendLogin should succeed")
	}
	ref := []byte{
		16, 1, 0, 223, 0, 0, 1, 0, 215, 0, 0, 0, 4, 0, 0, 116, 0, 16, 0, 0, 0, 1,
		6, 1, 100, 0, 0, 0, 0, 0, 0, 0, 224, 0, 0, 24, 16, 255, 255, 255, 4, 2, 0,
		0, 94, 0, 7, 0, 108, 0, 0, 0, 108, 0, 0, 0, 108, 0, 7, 0, 122, 0, 10, 0, 176,
		0, 4, 0, 142, 0, 7, 0, 156, 0, 2, 0, 160, 0, 8, 0, 18, 52, 86, 120, 144, 171,
		176, 0, 0, 0, 176, 0, 0, 0, 176, 0, 0, 0, 0, 0, 0, 0, 115, 0, 117, 0, 98,
		0, 100, 0, 101, 0, 118, 0, 49, 0, 97, 0, 112, 0, 112, 0, 110, 0, 97, 0,
		109, 0, 101, 0, 115, 0, 101, 0, 114, 0, 118, 0, 101, 0, 114, 0, 110, 0, 97,
		0, 109, 0, 101, 0, 108, 0, 105, 0, 98, 0, 114, 0, 97, 0, 114, 0, 121, 0, 101,
		0, 110, 0, 100, 0, 97, 0, 116, 0, 97, 0, 98, 0, 97, 0, 115, 0, 101, 0, 180, 0,
		0, 0, 2, 29, 0, 0, 0, 2, 24, 0, 0, 0, 102, 0, 101, 0, 100, 0, 97, 0, 117, 0,
		116, 0, 104, 0, 116, 0, 111, 0, 107, 0, 101, 0, 110, 0, 255}
	out := memBuf.Bytes()
	if !bytes.Equal(ref, out) {
		t.Log("Expected:")
		t.Log(hex.Dump(ref))
		t.Log("Returned:")
		t.Log(hex.Dump(out))
		t.Fatal("input output don't match")
	}
}

func TestSendSqlBatch(t *testing.T) {
	checkConnStr(t)
	p, err := parseConnectParams(makeConnStr(t).String())
	if err != nil {
		t.Error("parseConnectParams failed:", err.Error())
		return
	}

	conn, err := connect(context.Background(), nil, optionalLogger{testLogger{t}}, p)
	if err != nil {
		t.Error("Open connection failed:", err.Error())
		return
	}
	defer conn.buf.transport.Close()

	headers := []headerStruct{
		{hdrtype: dataStmHdrTransDescr,
			data: transDescrHdr{0, 1}.pack()},
	}
	err = sendSqlBatch72(conn.buf, "select 1", headers, true)
	if err != nil {
		t.Error("Sending sql batch failed", err.Error())
		return
	}

	ch := make(chan tokenStruct, 5)
	go processResponse(context.Background(), conn, ch, nil)
	defer func() {
		// make share ch is closed
		for range ch {
		}
	}()

	var lastRow []interface{}
loop:
	for tok := range ch {
		switch token := tok.(type) {
		case doneStruct:
			break loop
		case []columnStruct:
			conn.columns = token
		case []interface{}:
			lastRow = token
		default:
			t.Log("unknown token", tok)
		}
	}

	if len(lastRow) == 0 {
		t.Fatal("expected row but no row set")
	}

	switch value := lastRow[0].(type) {
	case int32:
		if value != 1 {
			t.Error("Invalid value returned, should be 1", value)
			return
		}
	}
}

func checkConnStr(t testing.TB) {
	if len(os.Getenv("SQLSERVER_DSN")) > 0 {
		return
	}
	if len(os.Getenv("HOST")) > 0 && len(os.Getenv("DATABASE")) > 0 {
		return
	}
	t.Skip("no database connection string")
}

// makeConnStr returns a URL struct so it may be modified by various
// tests before used as a DSN.
func makeConnStr(t testing.TB) *url.URL {
	return testConnParams(t).toUrl()
}

type testLogger struct {
	t testing.TB
}

func (l testLogger) Printf(format string, v ...interface{}) {
	l.t.Logf(format, v...)
}

func (l testLogger) Println(v ...interface{}) {
	l.t.Log(v...)
}

func testConnection(t *testing.T, connStr string) {
	conn, err := sql.Open("mssql", connStr)
	if err != nil {
		t.Fatal("Open connection failed:", err.Error())
		return
	}
	defer conn.Close()
	row := conn.QueryRow("select 1")
	var val int
	err = row.Scan(&val)
	if err != nil {
		t.Fatal("Scan failed:", err.Error())
	}
}

func TestConnect(t *testing.T) {
	checkConnStr(t)
	SetLogger(testLogger{t})
	testConnection(t, os.Getenv("SQLSERVER_DSN"))
}

func TestConnectViaIp(t *testing.T) {
	params := testConnParams(t)
	ips, err := net.LookupIP(params.host)
	if err != nil {
		t.Fatal("Unable to lookup IP", err)
	}
	params.host = ips[0].String()
	testConnection(t, params.toUrl().String())
}

func simpleQuery(conn *sql.DB, t *testing.T) (stmt *sql.Stmt) {
	stmt, err := conn.Prepare("select 1 as a")
	if err != nil {
		t.Error("Prepare failed:", err.Error())
		return nil
	}
	return stmt
}

func checkSimpleQuery(rows *sql.Rows, t *testing.T) {
	numrows := 0
	for rows.Next() {
		var val int
		err := rows.Scan(&val)
		if err != nil {
			t.Error("Scan failed:", err.Error())
		}
		if val != 1 {
			t.Error("query should return 1")
		}
		numrows++
	}
	if numrows != 1 {
		t.Error("query should return 1 row, returned", numrows)
	}
}

func TestQuery(t *testing.T) {
	conn := open(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	stmt := simpleQuery(conn, t)
	if stmt == nil {
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		t.Error("Query failed:", err.Error())
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		t.Error("getting columns failed", err.Error())
	}
	if len(columns) != 1 && columns[0] != "a" {
		t.Error("returned incorrect columns (expected ['a']):", columns)
	}

	checkSimpleQuery(rows, t)
}

func TestMultipleQueriesSequentialy(t *testing.T) {

	conn := open(t)
	defer conn.Close()

	stmt, err := conn.Prepare("select 1 as a")
	if err != nil {
		t.Error("Prepare failed:", err.Error())
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		t.Error("Query failed:", err.Error())
		return
	}
	defer rows.Close()
	checkSimpleQuery(rows, t)

	rows, err = stmt.Query()
	if err != nil {
		t.Error("Query failed:", err.Error())
		return
	}
	defer rows.Close()
	checkSimpleQuery(rows, t)
}

func TestMultipleQueryClose(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	stmt, err := conn.Prepare("select 1 as a")
	if err != nil {
		t.Error("Prepare failed:", err.Error())
		return
	}
	defer stmt.Close()

	func() {
		rows, err := stmt.Query()
		if err != nil {
			t.Fatal("Query failed:", err.Error())
		}
		defer rows.Close()
	}()

	func() {
		rows, err := stmt.Query()
		if err != nil {
			t.Fatal("Query failed:", err.Error())
		}
		defer rows.Close()
		checkSimpleQuery(rows, t)
	}()
}

func TestPing(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	conn.Ping()
}

func TestSecureWithInvalidHostName(t *testing.T) {
	checkConnStr(t)
	SetLogger(testLogger{t})

	dsn := makeConnStr(t)
	dsnParams := dsn.Query()
	dsnParams.Set("encrypt", "true")
	dsnParams.Set("TrustServerCertificate", "false")
	dsnParams.Set("hostNameInCertificate", "foo.bar")
	dsn.RawQuery = dsnParams.Encode()

	conn, err := sql.Open("mssql", dsn.String())
	if err != nil {
		t.Fatal("Open connection failed:", err.Error())
	}
	defer conn.Close()
	err = conn.Ping()
	if err == nil {
		t.Fatal("Connected to fake foo.bar server")
	}
}

func TestSecureConnection(t *testing.T) {
	checkConnStr(t)
	SetLogger(testLogger{t})

	dsn := makeConnStr(t)
	dsnParams := dsn.Query()
	dsnParams.Set("encrypt", "true")
	dsnParams.Set("TrustServerCertificate", "true")
	dsn.RawQuery = dsnParams.Encode()

	conn, err := sql.Open("mssql", dsn.String())
	if err != nil {
		t.Fatal("Open connection failed:", err.Error())
	}
	defer conn.Close()
	var msg string
	err = conn.QueryRow("select 'secret'").Scan(&msg)
	if err != nil {
		t.Fatal("cannot scan value", err)
	}
	if msg != "secret" {
		t.Fatal("expected secret, got: ", msg)
	}
	var secure bool
	err = conn.QueryRow("select encrypt_option from sys.dm_exec_connections where session_id=@@SPID").Scan(&secure)
	if err != nil {
		t.Fatal("cannot scan value", err)
	}
	if !secure {
		t.Fatal("connection is not encrypted")
	}
}

func TestBadConnect(t *testing.T) {
	checkConnStr(t)
	SetLogger(testLogger{t})
	connURL := makeConnStr(t)
	connURL.User = url.UserPassword("baduser", "badpwd")
	badDSN := connURL.String()

	conn, err := sql.Open("mssql", badDSN)
	if err != nil {
		t.Error("Open connection failed:", err.Error())
	}
	defer conn.Close()

	err = conn.Ping()
	if err == nil {
		t.Error("Ping should fail for connection: ", badDSN)
	}
}

func TestSSPIAuth(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Only on windows")
	}
	checkConnStr(t)
	connStr := makeConnStr(t)
	params := connStr.Query()
	params.Set("Integrated Security", "sspi")
	connStr.RawQuery = params.Encode()

	db, err := sql.Open("mssql", connStr.String())
	if err != nil {
		t.Error("Open failed", err)
	}
	defer db.Close()

	row := db.QueryRow("select 1, 'abc'")

	var somenumber int64
	var somechars string
	err = row.Scan(&somenumber, &somechars)
	if err != nil {
		t.Error("scan failed", err)
	}
	if somenumber != int64(1) || somechars != "abc" {
		t.Errorf("Invalid values from query: want {%d,'%s'}, got {%d,'%s'}", int64(1), "abc", somenumber, somechars)
	}
}

func TestUcs22Str(t *testing.T) {
	// Test valid input
	s, err := ucs22str([]byte{0x31, 0, 0x32, 0, 0x33, 0}) // 123 in UCS2 encoding
	if err != nil {
		t.Errorf("ucs22str should not fail for valid ucs2 byte sequence: %s", err)
	}
	if s != "123" {
		t.Errorf("ucs22str expected to return 123 but it returned %s", s)
	}

	// Test invalid input
	_, err = ucs22str([]byte{0})
	if err == nil {
		t.Error("ucs22str should fail on single byte input, but it didn't")
	}
}

func TestReadUcs2(t *testing.T) {
	buf := bytes.NewBuffer([]byte{0x31, 0, 0x32, 0, 0x33, 0}) // 123 in UCS2 encoding
	s, err := readUcs2(buf, 3)
	if err != nil {
		t.Errorf("readUcs2 should not fail for valid ucs2 byte sequence: %s", err)
	}
	if s != "123" {
		t.Errorf("readUcs2 expected to return 123 but it returned %s", s)
	}

	buf = bytes.NewBuffer([]byte{0})
	_, err = readUcs2(buf, 1)
	if err == nil {
		t.Error("readUcs2 should fail on single byte input, but it didn't")
	}
}

func TestReadUsVarChar(t *testing.T) {
	// should succeed for valid buffer
	buf := bytes.NewBuffer([]byte{3, 0, 0x31, 0, 0x32, 0, 0x33, 0}) // 123 in UCS2 encoding with length prefix 3 uint16
	s, err := readUsVarChar(buf)
	if err != nil {
		t.Errorf("readUsVarChar should not fail for valid ucs2 byte sequence: %s", err)
	}
	if s != "123" {
		t.Errorf("readUsVarChar expected to return 123 but it returned %s", s)
	}

	// should fail for empty buffer
	buf = bytes.NewBuffer([]byte{})
	_, err = readUsVarChar(buf)
	if err == nil {
		t.Error("readUsVarChar should fail on empty buffer, but it didn't")
	}
}

func TestReadBVarByte(t *testing.T) {
	memBuf := bytes.NewBuffer([]byte{3, 1, 2, 3})
	s, err := readBVarByte(memBuf)
	if err != nil {
		t.Errorf("readUsVarByte should not fail for valid buffer: %s", err)
	}
	if !bytes.Equal(s, []byte{1, 2, 3}) {
		t.Errorf("readUsVarByte expected to return [1 2 3] but it returned %v", s)
	}

	// test empty buffer
	memBuf = bytes.NewBuffer([]byte{})
	s, err = readBVarByte(memBuf)
	if err == nil {
		t.Error("readUsVarByte should fail on empty buffer, but it didn't")
	}

	// test short buffer
	memBuf = bytes.NewBuffer([]byte{1})
	s, err = readBVarByte(memBuf)
	if err == nil {
		t.Error("readUsVarByte should fail on short buffer, but it didn't")
	}
}

func BenchmarkPacketSize(b *testing.B) {
	checkConnStr(b)
	p, err := parseConnectParams(makeConnStr(b).String())
	if err != nil {
		b.Error("parseConnectParams failed:", err.Error())
		return
	}

	benchmarks := []struct {
		name       string
		packetSize uint16
	}{
		{name: "PacketSize 2048", packetSize: 2048},
		{name: "PacketSize 4096", packetSize: 4096},
		{name: "PacketSize 8192", packetSize: 8192},
		{name: "PacketSize 16384", packetSize: 16384},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				p.packetSize = bm.packetSize
				runBatch(b, p)
			}
		})
	}

}

func runBatch(t testing.TB, p connectParams) {
	conn, err := connect(context.Background(), nil, optionalLogger{testLogger{t}}, p)
	if err != nil {
		t.Error("Open connection failed:", err.Error())
		return
	}
	defer conn.buf.transport.Close()

	headers := []headerStruct{
		{hdrtype: dataStmHdrTransDescr,
			data: transDescrHdr{0, 1}.pack()},
	}
	err = sendSqlBatch72(conn.buf, "select 1", headers, true)
	if err != nil {
		t.Error("Sending sql batch failed", err.Error())
		return
	}

	ch := make(chan tokenStruct, 5)
	go processResponse(context.Background(), conn, ch, nil)

	var lastRow []interface{}
loop:
	for tok := range ch {
		switch token := tok.(type) {
		case doneStruct:
			break loop
		case []columnStruct:
			conn.columns = token
		case []interface{}:
			lastRow = token
		default:
			t.Log("unknown token", tok)
		}
	}

	if len(lastRow) == 0 {
		t.Fatal("expected row but no row set")
	}

	switch value := lastRow[0].(type) {
	case int32:
		if value != 1 {
			t.Error("Invalid value returned, should be 1", value)
			return
		}
	}
}
