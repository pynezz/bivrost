package ipcserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/pynezz/bivrost/internal/database"
	"github.com/pynezz/bivrost/internal/database/models"
	"github.com/pynezz/bivrost/internal/database/stores"
	"github.com/pynezz/bivrost/internal/ipc"
	"github.com/pynezz/bivrost/modules"

	util "github.com/pynezz/pynezzentials"
	"github.com/pynezz/pynezzentials/ansi"
	"github.com/pynezz/pynezzentials/fsutil"
)

/* CONSTANTS
 * The following constants are used for the IPC communication between the connector and the other modules.
 */
const (
	AF_UNIX  = "unix"     // UNIX domain sockets
	AF_DGRAM = "unixgram" // UNIX domain datagram sockets as specified in net package
)

/* IDENTIFIERS
 * To identify the module, the client will send a 4 byte identifier as part of the header.
 */
var MODULEIDENTIFIERS map[string][]byte
var MSGTYPE map[string]byte
var SERVERIDENTIFIER [4]byte
var IPCID []byte

/* TYPES
 * Types for the IPC communication between the connector and the other modules.
 */
type IPCServer struct {
	path         string
	identifier   string
	conn         net.Listener
	moduleStates map[string]*ipc.ModuleState
}

func (s *IPCServer) CloseConn() {
	s.conn.Close()
}

func init() {
	MODULEIDENTIFIERS = map[string][]byte{}
}

// LoadModules loads the modules from the given file
// It will parse the config file and add the modules to the server map
func LoadModules(path string) {
	if !fsutil.FileExists(path) {
		ansi.PrintError("LoadModules(): File does not exist: " + path)
	}
	ansi.PrintSuccess("File exists: " + path)
	f, err := os.Open(path)
	if err != nil {
		ansi.PrintError("LoadModules(): " + err.Error())
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		parts := strings.Split(line, " ")
		if len(parts[0]) < 1 {
			// empty line
			continue
		}
		firstChar := parts[0][0]
		if firstChar == '#' || firstChar == ' ' || firstChar == '\t' || firstChar == '/' || firstChar == '*' || firstChar == '\n' || firstChar == '\r' {
			// comment
			continue
		}

		for i, part := range parts {
			if part == "\t" || part == " " {
				// skip
				continue
			}
			parts[i] = strings.TrimSpace(part)
		}

		AddModule(parts[0], []byte(parts[1])) // Add module to the server
		ansi.PrintColorf(ansi.LightCyan, "Loaded module: %s", parts[0])
	}
}

// NewIPCServer creates a new IPC server and returns it.
func NewIPCServer(name string, identifier string) *IPCServer {
	path := ipc.DefaultSock(name)
	IPCID = []byte(identifier)
	ipc.SetIPCID(IPCID)
	SetServerIdentifier(IPCID)

	moduleStates := make(map[string]*ipc.ModuleState)

	ansi.PrintColorf(ansi.LightCyan, "[SOCKETS] IPC server path: %s", path)

	return &IPCServer{
		path:         path,
		identifier:   identifier,
		conn:         nil,
		moduleStates: moduleStates,
	}
}

// Add a new module identifier to the map
func AddModule(identifier string, id []byte) {
	if len(id) > 4 {
		ansi.PrintError("AddModule(): Identifier length must be 4 bytes")
		ansi.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	MODULEIDENTIFIERS[identifier] = id

	ansi.PrintSuccess("added module:" + string(MODULEIDENTIFIERS[identifier]))
}

// Set the server identifier to the SERVERIDENTIFIER variable
func SetServerIdentifier(id []byte) {
	if len(id) > 4 {
		ansi.PrintError("SetServerIdentifier(): Identifier length must be 4 bytes")
		ansi.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	SERVERIDENTIFIER = [4]byte(id) // Convert the slice to an array
	ansi.PrintSuccess("Set server identifier: " + string(SERVERIDENTIFIER[:]))
}

// Write a socket file and add it to the map
func (s *IPCServer) InitServerSocket() bool {
	// Making sure the socket is clean before starting
	if err := os.RemoveAll(s.path); err != nil {
		ansi.PrintError("InitServerSocket(): Failed to remove old socket: " + err.Error())
		return false
	}
	return true
}

// Creates a new listener on the socket path (which should be set in the config in the future)
func (s *IPCServer) Listen() {
	ansi.PrintColorBold(ansi.DarkGreen, "🎉 IPC server running!")
	var err error
	s.conn, err = net.Listen(AF_UNIX, s.path)
	if err != nil {
		ansi.PrintError("Listen(): " + err.Error())
		return
	}
	ansi.PrintColorf(ansi.LightCyan, "[SOCKETS] Starting listener on %s", s.path)

	for {
		ansi.PrintDebug("Waiting for connection...")
		ansi.PrintDebug("Network: " + s.conn.Addr().Network())

		conn, err := s.conn.Accept()
		ansi.PrintColorf(ansi.LightCyan, "[SOCKETS]: New connection from %s", conn.LocalAddr().String())

		if err != nil {
			ansi.PrintError("Listen(): " + err.Error())
			continue
		}
		go s.handleConnection(conn) // Handle the connection in a separate goroutine to handle multiple connections
	}
}

// Function to create a new IPCMessage based on the identifier key
func NewIPCMessage(identifierKey string, messageType byte, data []byte) (*ipc.IPCRequest, error) {
	identifier := modules.Mids.GetModuleIdentifier(identifierKey)
	ansi.PrintDebug("NewIPCMessage from module with key: " + identifierKey)

	messageId := fmt.Sprintf("%d-%s", time.Now().UnixNano(), identifierKey)

	var id [4]byte
	copy(id[:], identifier[:4]) // Ensure no out of bounds panic

	message := ipc.IPCMessage{
		Data:       data,
		StringData: string(data),
	}

	return &ipc.IPCRequest{
		MessageSignature: []byte(messageId),
		Header: ipc.IPCHeader{
			Identifier:  id,
			MessageType: messageType,
		},
		Message:    message,
		Timestamp:  util.UnixNanoTimestamp(),
		Checksum32: int(crc(data)),
	}, nil
}

// Return the parsed IPCRequest object
func parseConnection(c net.Conn) (ipc.IPCRequest, error) {
	var request ipc.IPCRequest
	decoder := gob.NewDecoder(c)
	err := decoder.Decode(&request)
	return request, err
}

// Parse the metadata from the message
// Returns the source, destination, and a boolean indicating if the metadata is nil
func parseMetadata(msg ipc.Metadata) (ipc.Metadata, bool) {

	metadata := msg

	v := ""
	if metadata.Method == "POST" {
		v = "send data to"
	} else if metadata.Method == "GET" {
		v = "get data from"
	} else if metadata.Method == "PUT" {
		v = "update data in"
	} else if metadata.Method == "DELETE" {
		v = "delete data from"
	} else {
		v = "???"
	}

	sentence := fmt.Sprintf("\n %s wants to %s %s with id %s \n", metadata.Source, v, metadata.Destination.Object, metadata.Destination.Object.Id)
	ansi.PrintBold(sentence)
	ansi.PrintItalic("Database name: " + msg.Destination.Object.Database.Name + "\nTable name: " + metadata.Destination.Object.Database.Table)

	return metadata, true
}

type JsonResponse struct {
	Metadata ipc.Metadata `json:"metadata"`
	Data     interface{}  `json:"data"` //? Wait, I thought this was supposed to be an array of AttackDetail..?
	Model    string       `json:"model"`
}

func parseData(msg *ipc.IPCMessage) (ipc.GenericData, JsonResponse) {
	var data ipc.GenericData
	var jsonData JsonResponse

	switch msg.Datatype {
	case ipc.DATA_TEXT:
		fmt.Println("Data is string")
	case ipc.DATA_INT:
		// Might be a disconnect message
		fmt.Println("Data is integer")
	case ipc.DATA_JSON:
		// Parse the JSON data
		fmt.Println("Data is json / generic data")

		err := json.Unmarshal(msg.Data, &jsonData)
		if err != nil {
			errMsg := fmt.Errorf("error unmarshaling JSON data: %w", err)
			fmt.Println(errMsg)
		}
		return data, jsonData

	case ipc.DATA_YAML:
		// Parse the YAML data
		fmt.Println("Data is YAML / generic data")
		err := yaml.Unmarshal(msg.Data, &data)
		if err != nil {
			fmt.Println("Error unmarshaling YAML data:", err)
		}
		handleGenericData(msg.Data)

	case ipc.DATA_BIN:
		// Parse the binary data
		fmt.Println("Data is binary / generic data")
		handleGenericData(data)

	default:
		// Default to generic data
		fmt.Println("Data is generic")
		gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(&data)
		handleGenericData(data)

	}

	if data == nil {
		fmt.Println("Data is nil")
	}

	return data, jsonData
}

func handleGenericData( /* t reflect.Type, */ dataType any) {
	// Handle the data
	fmt.Println("Handling generic data...")

	dataBytes, ok := dataType.([]byte)
	if !ok {
		fmt.Println("Invalid data type")
		return
	}

	gob.NewDecoder(bytes.NewReader(dataBytes)).Decode(&dataType)
	fmt.Println("Data: ", dataType)
}

// handleConnection handles the incoming connection
func (s *IPCServer) handleConnection(c net.Conn) {
	// defer c.Close()

	ansi.PrintColorf(ansi.LightCyan, "[SOCKETS] Handling connection...")

	for {
		inboundRequest, err := parseConnection(c)
		if err != nil {
			if err == io.EOF {
				ansi.PrintDebug("Connection closed by client")
				break
			}
			fmt.Println(ansi.Errorf("Error parsing request: " + err.Error()))
			break
		}

		_, d := parseData(&inboundRequest.Message) // Should be of type ipc.IPCMessage
		if d == *new(JsonResponse) {
			fmt.Println("Data is nil")
			return
		}

		moduleName := modules.Mids.GetModuleName(inboundRequest.Header.Identifier)
		if moduleName == "" {
			ansi.PrintColorf(ansi.LightRed, "[+] Added module name: %s", moduleName)
			modules.Mids.StoreModuleIdentifier(string(inboundRequest.Header.Identifier[:]), inboundRequest.Header.Identifier)
		} else {
			ansi.PrintColorf(ansi.LightCyan, "Module name: %s", moduleName)
		}

		source := fmt.Sprint(modules.Mids.GetModuleIdentifier(moduleName)) // Sorry about this (this should print the identifier of the module that sent the request)
		fmt.Println("Source: " + source)

		var response []byte

		// Parse metadata
		mData, ok := parseMetadata(d.Metadata)
		if !ok {
			ansi.PrintWarning("Metadata is nil")
		} else {
			ansi.PrintSuccess("Metadata: " + fmt.Sprintf("%v", mData))

			// If there is data to fetch, fetch it
			if mData.Method == "GET" {

				ansi.PrintBold("Got a GET request - fetching data...")

				// Get the data source
				databaseName := mData.Destination.Object.Database.Name
				tableName := mData.Destination.Object.Database.Table

				// Get the data sources
				ansi.PrintColorf(ansi.LightCyan, "Database: %s\nTable: %s", databaseName, tableName)

				// Ask database for the data
				// TODO: Remember to make sure only the latest data is fetched
				// ctx := context.WithValue(*s.moduleStates, "tableName", 0) // TODO: <- Add row id here
				ctx := context.WithValue(context.Background(), tableToModel(tableName), s.moduleStates)
				val := ctx.Value(tableToModel(tableName))

				fmt.Scanln("[⏸️] context value: ", val)

				latestData := fetchLatestLogData(ctx, tableName)

				if latestData.data == nil {
					ansi.PrintError("Source not found")
					response = []byte("Error, source not found")
					// Respond with an error
					ansi.PrintError("Failed to fetch the data")

				} else {
					// Get the logs
					response = latestData.data
					// Respond with the data
					ansi.PrintSuccess("Data fetched successfully")
				}
			}
			if mData.Method == "POST" {
				s.handlePost(mData, d.Data)
				response = []byte("OK")
			}
		}

		if err := reply(response, c, inboundRequest, s); err != nil {
			ansi.PrintError("handleConnection: " + err.Error())
			return
		}
	}
}

func reply(response []byte, c net.Conn, inboundRequest ipc.IPCRequest, s *IPCServer) error {
	// Finally, respond to the client
	var err error

	moduleId := string(inboundRequest.Header.Identifier[:]) // TODO: req.Header.Identifier is the server identifier, not the module identifier
	if crcOk := s.CheckCRC32(inboundRequest); !crcOk {
		response = []byte("CHKSUM ERROR")
		ansi.PrintError("Checksum error")
		return fmt.Errorf(string(response))
	}

	err = s.respond(c, response, moduleId)
	if err != nil {
		ansi.PrintError("handleConnection: " + err.Error())
		return err
	}
	responseTime(inboundRequest.Timestamp)

	return nil
}

// Get the model based on the table name
func tableToModel(tableName string) any {
	for range models.GetModels() {
		switch tableName {
		case models.ATTACK_TYPE:
			return models.AttackType{}
		case models.NGINX_LOGS:
			return models.NginxLog{}
		case models.SYN_TRAFFIC:
			return models.SynTraffic{}
		case models.GEO_DATA:
			return models.GeoData{}
		case models.GEO_LOCATION_DATA:
			return models.GeoLocationData{}
		case models.THREAT_RECORDS:
			return models.ThreatRecord{}
		default:
			return nil
		}
	}
	return nil
}

func (s *IPCServer) handlePost(m ipc.Metadata, data interface{}) {
	ansi.PrintBold("Got a POST request - inserting data...")

	databaseName := m.Destination.Object.Database.Name
	tableName := m.Destination.Object.Database.Table

	bytes, err := json.Marshal(data)
	if err != nil {
		ansi.PrintError("Failed to marshal the data: " + err.Error())
		return
	}
	if len(bytes) > 100 {
		fmt.Println("Data: ", string(bytes[:100]))
	} else {
		fmt.Println("Data: ", string(bytes))
	}

	ansi.PrintDebug("Key value pairs in the data object: ")
	// field := 0
	fields := 0

	if dataList, ok := data.([]interface{}); ok {
		fields += len(dataList)
		// for _, v := range dataList {
		// 	field++
		// fmt.Println(field, "=>", v)
		// Print truncated data:
		// if len(v.(map[string]interface{})) > 50 {
		// fmt.Println(field, "=>", string(bytes[0:50]))
		// } else {
		// fmt.Println(field, "=>", string(bytes))
		// }
		// }
	}
	fmt.Println("Field count: ", fields)

	marshalled, err := json.Marshal(data)
	if err != nil {
		ansi.PrintError("Failed to marshal the data: " + err.Error())
	}

	switch tableName {
	case models.ATTACK_TYPE:
		// Insert the data into the database
		ansi.PrintDebug("Inserting data into the database...")
		var tmpData []models.AttackType

		json.Unmarshal(marshalled, &tmpData)
		insertData[models.AttackType](databaseName, tableName, tmpData)
	case models.INDICATORS_LOG:
		// Insert the data into the database
		var tmpData []models.IndicatorsLog
		json.Unmarshal(marshalled, &tmpData)
		insertData[models.IndicatorsLog](databaseName, tableName, tmpData)
	case models.NGINX_LOGS:
		// Insert the data into the database
		var tmpData []models.NginxLog
		json.Unmarshal(marshalled, &tmpData)
		insertData[models.NginxLog](databaseName, tableName, tmpData)
	case models.SYN_TRAFFIC:
		// Insert the data into the database
		var tmpData []models.SynTraffic

		json.Unmarshal(marshalled, &tmpData)
		insertData[models.SynTraffic](databaseName, tableName, tmpData)
	case models.GEO_DATA:
		// Insert the data into the database
		var tmpData []models.GeoData
		json.Unmarshal(marshalled, &tmpData)
		insertData[models.GeoData](databaseName, tableName, tmpData)
	case models.GEO_LOCATION_DATA:
		// Insert the data into the database
		var tmpData []models.GeoLocationData
		json.Unmarshal(marshalled, &tmpData)
		insertData[models.GeoLocationData](databaseName, tableName, tmpData)
	case models.THREAT_RECORDS:
		var tmpData []models.ThreatRecord
		json.Unmarshal(marshalled, &tmpData)
		insertData[models.ThreatRecord](databaseName, tableName, tmpData)
	default:
		// Insert the data into the database
		ansi.PrintDebug("Unknown table name")
	}
}

// WIP: ✅ Tested - working again! 03.06.2024
func insertData[StoreType any](databaseName, tableName string, data any) {
	// Get the data store
	d := data.([]StoreType)

	s, err := stores.Use(tableName)
	if err != nil {

		s := &database.DataStore[StoreType]{}

		ansi.PrintError("Failed to get the data store: " + s.Name())
	}

	ansi.PrintDebug("Getting the data store...")

	// Depending on the table name, insert the data into the database
	switch tableName {
	case models.ATTACK_TYPE:
		// Insert the data into the database
		ansi.PrintDebug("Inserting data into the database...")
		var tmpData []models.AttackType
		tmpBytes, err := json.Marshal(d)
		if err != nil {
			ansi.PrintError("Failed to marshal the data: " + err.Error())
		}
		err = json.Unmarshal(tmpBytes, &tmpData)
		if err != nil {
			ansi.PrintError("Failed to unmarshal the data: " + err.Error())
		}

		// fmt.Printf("Data: %v\n", tmpData)
		logsChannel := make(chan models.AttackType)
		go func() {
			for _, log := range tmpData {
				logsChannel <- log
			}
			close(logsChannel)
		}()

		err = s.AttackTypeStore.InsertBulk(logsChannel, len(tmpData))
		if err != nil {
			ansi.PrintError("Failed to insert the data: " + err.Error())
		} else {
			ansi.PrintSuccess("Data inserted successfully")
		}

	case models.NGINX_LOGS:
		return
	case models.SYN_TRAFFIC:
		var tmpData []models.SynTraffic
		tmpBytes, err := json.Marshal(d)
		if err != nil {
			ansi.PrintError("Failed to marshal the data: " + err.Error())
		}
		err = json.Unmarshal(tmpBytes, &tmpData)
		if err != nil {
			ansi.PrintError("Failed to unmarshal the data: " + err.Error())
		}
		logsChannel := make(chan models.SynTraffic)
		go func() {
			for _, log := range tmpData {
				logsChannel <- log
			}
			close(logsChannel)

		}()

		err = s.SynTrafficStore.InsertBulk(logsChannel, len(tmpData))
		if err != nil {
			ansi.PrintError("Failed to insert the data: " + err.Error())
		} else {
			ansi.PrintSuccess("Data inserted successfully")
		}
	case models.GEO_DATA:
		var tmpData []models.GeoData
		tmpBytes, err := json.Marshal(d)
		if err != nil {
			ansi.PrintError("Failed to marshal the data: " + err.Error())
		}
		err = json.Unmarshal(tmpBytes, &tmpData)
		if err != nil {
			ansi.PrintError("Failed to unmarshal the data: " + err.Error())
		}

		logsChannel := make(chan models.GeoData)
		go func() {
			for _, log := range tmpData {
				logsChannel <- log
			}
			close(logsChannel)

		}()

		err = s.GeoDataStore.InsertBulk(logsChannel, len(tmpData))
		if err != nil {
			ansi.PrintError("Failed to insert the data: " + err.Error())
		} else {
			ansi.PrintSuccess("Data inserted successfully")
		}
	case models.GEO_LOCATION_DATA:
		var tmpData []models.GeoLocationData
		tmpBytes, err := json.Marshal(d)
		if err != nil {
			ansi.PrintError("Failed to marshal the data: " + err.Error())
		}
		err = json.Unmarshal(tmpBytes, &tmpData)
		if err != nil {
			ansi.PrintError("Failed to unmarshal the data: " + err.Error())
		}
		logsChannel := make(chan models.GeoLocationData)
		go func() {
			for _, log := range tmpData {
				logsChannel <- log
			}
			close(logsChannel)

		}()

		err = s.GeoLocationDataStore.InsertBulk(logsChannel, len(tmpData))
		if err != nil {
			ansi.PrintError("Failed to insert the data: " + err.Error())
		} else {
			ansi.PrintSuccess("Data inserted successfully")
		}
	case models.THREAT_RECORDS:
		var tmpData []models.ThreatRecord

		// Might be unneccessary at this stage due to refactoring made in handlePost()
		tmpBytes, err := json.Marshal(d)
		if err != nil {
			ansi.PrintError("failed to marshal the data " + err.Error())
		}
		err = json.Unmarshal(tmpBytes, &tmpData)
		if err != nil {
			ansi.PrintError("Failed to unmarshal the data: " + err.Error())
		}

		logsChannel := make(chan models.ThreatRecord)
		go func() {
			for _, log := range tmpData {
				logsChannel <- log
			}
			close(logsChannel)
		}()

		err = s.ThreatRecordStore.InsertBulk(logsChannel, len(tmpData))
		if err != nil {
			ansi.PrintError("failed to insert threat records data " + err.Error())
		} else {
			ansi.PrintSuccess("Successfully inserted threat records")
		}

	default:
		ansi.PrintError("Table not found: " + tableName)
	}
}

// Parsing into this object from a "LogEntry" struct
type LogData struct {
	lastRowID int
	data      []byte // LogEntry struct
}

// TODO: FORTSETT HER
func fetchLatestLogData(ctx context.Context, tableName string) LogData {
	moduleStatesMap, ok := ctx.Value(tableToModel(tableName)).(map[string]*ipc.ModuleState)
	if !ok {
		// Handle the case where the value is not a map[string]*ModuleState
		ansi.PrintDebug(" [modulestate] failed to get the module states map")
		fmt.Scanln("Press [Enter] to continue...")

		return LogData{}
	}

	moduleState, ok := moduleStatesMap[tableName]
	if !ok {
		// Create a new ModuleState for this table
		moduleState = &ipc.ModuleState{}
		moduleStatesMap[tableName] = moduleState
	}
	// Get the data from the database
	ansi.PrintDebug(" [modulestate] fetching the latest log data...")
	s, err := stores.Use(tableName)
	if err != nil {
		ansi.PrintError(" [modulestate] failed to get the data store: " + s.NginxLogStore.Name())
	}

	// Use a read lock to safely read the LastRowID
	moduleState.RLock()
	lastRowID := moduleState.LastRowID
	moduleState.RUnlock()

	ansi.PrintDebug("Getting the data store...")

	logs, err := s.NginxLogStore.GetLogRangeFromID(lastRowID) // The store is also an issue here....
	if err != nil {
		ansi.PrintError("[modulestate] failed to get all logs: " + err.Error())
	}

	ansi.PrintDebug("[modulestate] getting all logs starting with " + fmt.Sprintf("%d", lastRowID) + "...")
	fmt.Scanln("Press [Enter] to continue...")

	returnData := LogData{
		lastRowID: lastRowID,
		data:      nil,
	}

	returnData.data, err = json.Marshal(&logs)
	if err != nil {
		ansi.PrintError("Failed to marshal the logs: " + err.Error())
	}

	moduleState.Lock()
	moduleState.LastRowID += len(logs) // Also a bug was here (+ len(logs) instead of += len(logs))
	moduleState.Unlock()
	ansi.PrintDebug(" [modulestate] updated the last row ID to " + fmt.Sprintf("%d", moduleState.LastRowID))
	ansi.PrintDebug(" [modulestate] Returning the data...")
	fmt.Scanln("Press [Enter] to continue...")
	return returnData
}

func getResource(mc *modules.ModuleConfig) string {
	return mc.Database.Path
}

func (s *IPCServer) CheckCRC32(req ipc.IPCRequest) bool {
	return req.Checksum32 == int(crc(req.Message.Data))
}

// c is the connection to the client
func (s *IPCServer) respond(c net.Conn, data []byte, moduleId string) error {
	ansi.PrintDebug("Responding to the client...")

	var response *ipc.IPCRequest
	var err error

	response, err = NewIPCMessage(moduleId, ipc.MSG_ACK, data)
	if err != nil {
		return err
	}

	var responseBuffer bytes.Buffer
	encoder := gob.NewEncoder(&responseBuffer)
	err = encoder.Encode(response)
	if err != nil {
		return err
	}

	_, err = c.Write(responseBuffer.Bytes())
	if err != nil {
		return err
	}
	ansi.PrintColor(ansi.BgGreen, "🚀 Response sent!")

	return nil
}

// ReturnData is a struct that holds the metadata and the data
type ReturnData struct {
	Metadata ipc.Metadata `json:"metadata"`
	Data     interface{}  `json:"data"`
}

type DataSource struct {
}

// TODO: Where should we tell the server where to get the logs from?
func (d *ReturnData) GetLogs(path string, filter string) {
	// Read the logs from the file
	if fsutil.FileExists(path) {

	}

	// Not sure if this fits here. Might rather implement it in the internal package
}

// Calculate the response time
func responseTime(reqTime int64) {
	currTime := util.UnixNanoTimestamp()
	diff := currTime - reqTime
	fmt.Printf("Response time: %d\n", diff)
	fmt.Printf("Seconds: %f\n", float64(diff)/1e9)
	fmt.Printf("Milliseconds: %f\n", float64(diff)/1e6)
	fmt.Printf("Microseconds: %f\n", float64(diff)/1e3)
}

func NewIPCID(identifier string, id []byte) {
	if len(id) > 4 {
		ansi.PrintError("NewIPCID(): Identifier length must be 4 bytes")
		ansi.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	ipc.SetIPCID(id)
}

func Cleanup() {
	ansi.PrintItalic("\t... IPC server cleanup complete.")
}

func crc(b []byte) uint32 {
	return crc32.ChecksumIEEE(b)
}

// TODO's: 1]
// - check the config for activated modules
// - check if the module is already activated
// - check if connection with the module already exists
// - create socket names based on the module name
