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

	"gopkg.in/yaml.v3"

	"github.com/pynezz/bivrost/internal/database"
	"github.com/pynezz/bivrost/internal/database/models"
	"github.com/pynezz/bivrost/internal/database/stores"
	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/ipc"
	"github.com/pynezz/bivrost/internal/util"
	"github.com/pynezz/bivrost/modules"
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
		util.PrintError("LoadModules(): File does not exist: " + path)
	}
	util.PrintSuccess("File exists: " + path)
	f, err := os.Open(path)
	if err != nil {
		util.PrintError("LoadModules(): " + err.Error())
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
		util.PrintColorf(util.LightCyan, "Loaded module: %s", parts[0])
	}
}

// NewIPCServer creates a new IPC server and returns it.
func NewIPCServer(name string, identifier string) *IPCServer {
	path := ipc.DefaultSock(name)
	IPCID = []byte(identifier)
	ipc.SetIPCID(IPCID)
	SetServerIdentifier(IPCID)

	moduleStates := make(map[string]*ipc.ModuleState)

	util.PrintColorf(util.LightCyan, "[SOCKETS] IPC server path: %s", path)

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
		util.PrintError("AddModule(): Identifier length must be 4 bytes")
		util.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	MODULEIDENTIFIERS[identifier] = id

	util.PrintSuccess("added module:" + string(MODULEIDENTIFIERS[identifier]))
}

// Set the server identifier to the SERVERIDENTIFIER variable
func SetServerIdentifier(id []byte) {
	if len(id) > 4 {
		util.PrintError("SetServerIdentifier(): Identifier length must be 4 bytes")
		util.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	SERVERIDENTIFIER = [4]byte(id) // Convert the slice to an array
	util.PrintSuccess("Set server identifier: " + string(SERVERIDENTIFIER[:]))
}

// Write a socket file and add it to the map
func (s *IPCServer) InitServerSocket() bool {
	// Making sure the socket is clean before starting
	if err := os.RemoveAll(s.path); err != nil {
		util.PrintError("InitServerSocket(): Failed to remove old socket: " + err.Error())
		return false
	}
	return true
}

// Creates a new listener on the socket path (which should be set in the config in the future)
func (s *IPCServer) Listen() {
	util.PrintColorBold(util.DarkGreen, "🎉 IPC server running!")
	var err error
	s.conn, err = net.Listen(AF_UNIX, s.path)
	if err != nil {
		util.PrintError("Listen(): " + err.Error())
		return
	}
	util.PrintColorf(util.LightCyan, "[SOCKETS] Starting listener on %s", s.path)

	for {
		util.PrintDebug("Waiting for connection...")
		util.PrintDebug("Network: " + s.conn.Addr().Network())

		conn, err := s.conn.Accept()
		util.PrintColorf(util.LightCyan, "[SOCKETS]: New connection from %s", conn.LocalAddr().String())

		if err != nil {
			util.PrintError("Listen(): " + err.Error())
			continue
		}
		go s.handleConnection(conn) // Handle the connection in a separate goroutine to handle multiple connections
	}
}

// Function to create a new IPCMessage based on the identifier key
func NewIPCMessage(identifierKey string, messageType byte, data []byte) (*ipc.IPCRequest, error) {
	identifier := modules.Mids.GetModuleIdentifier(identifierKey)
	util.PrintDebug("NewIPCMessage from module with key: " + identifierKey)

	var id [4]byte
	copy(id[:], identifier[:4]) // Ensure no out of bounds panic

	message := ipc.IPCMessage{
		Data:       data,
		StringData: string(data),
	}

	return &ipc.IPCRequest{
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
	// var reqBuffer bytes.Buffer
	util.PrintDebug("Trying to decode the bytes to a request struct...")
	util.PrintColorf(util.LightCyan, "Decoding the bytes to a request struct... %v", c)

	decoder := gob.NewDecoder(c)
	err := decoder.Decode(&request)
	if err != nil {
		util.PrintWarning("parseConnection: Error decoding the request: \n > " + err.Error())
		return request, err
	}

	fmt.Printf("Message ID: %s\n", string(request.MessageSignature))

	return request, nil
}

// Parse the metadata from the message
// Returns the source, destination, and a boolean indicating if the metadata is nil
func parseMetadata(msg ipc.Metadata) (ipc.Metadata, bool) {

	// TODO: Might be reasonable to implement the Metadata struct here (ipc.Metadata)
	metadata := msg

	// source := metadata.(map[string]interface{})["source"].(string)
	// destination := metadata.(map[string]interface{})["destination"].(map[string]interface{})["destination"]
	// destinationId := destination.(map[string]interface{})["id"].(string)
	// destinationName := destination.(map[string]interface{})["name"]
	// destinationInfo := destination.(map[string]interface{})["database"]

	// databaseName := destinationInfo.(map[string]interface{})["name"]
	// tableName := destinationInfo.(map[string]interface{})["table"]

	// m := ipc.Metadata{
	// 	Source: source,
	// 	Destination: ipc.Destination{
	// 		Object: ipc.Object{
	// 			Id:   destinationId,
	// 			Name: destinationName.(string),
	// 			Database: ipc.Database{
	// 				Name:  databaseName.(string),
	// 				Table: tableName.(string),
	// 			},
	// 		},
	// 	},
	// 	Method: msg.Method,
	// }

	// method := metadata.(map[string]interface{})["method"]

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

	util.PrintColorf(util.DarkYellow, "Metadata object: %s", metadata)

	sentence := fmt.Sprintf("\n %s wants to %s %s with id %s \n", metadata.Source, v, metadata.Destination.Object, metadata.Destination.Object.Id)
	util.PrintBold(sentence)
	util.PrintItalic("Database name: " + msg.Destination.Object.Database.Name + "\nTable name: " + metadata.Destination.Object.Database.Table)

	return metadata, true
}

// Parse the method/verb from the message
func parseVerb(msg ipc.GenericData) string {
	v := msg["metadata"].(map[string]interface{})["method"].(string)
	if v == "" {
		return "nil"
	}
	util.PrintSuccess("Verb: " + v)
	return v
}

func getModel(tableName string) any {
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
		default:
			return nil
		}
	}
	return nil
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

		// getModel(msg.Data["Model"].(string))
		// Get the model based on the table name
		// handleGenericData(msg.Data)

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

func handleGenericData(dataType any) {
	// Handle the data
	fmt.Println("Handling generic data...")

}

// handleConnection handles the incoming connection
func (s *IPCServer) handleConnection(c net.Conn) {
	// defer c.Close()

	util.PrintColorf(util.LightCyan, "[SOCKETS] Handling connection...")

	for {
		inboundRequest, err := parseConnection(c)
		if err != nil {
			if err == io.EOF {
				util.PrintDebug("Connection closed by client")
				break
			}
			util.PrintError("Error parsing request: " + err.Error())
			break
		}

		_, d := parseData(&inboundRequest.Message) // Should be of type ipc.IPCMessage
		if len(d.Data.(map[string]interface{})) <= 0 {
			fmt.Println("Data is nil")
			return
		}

		moduleName := modules.Mids.GetModuleName(inboundRequest.Header.Identifier)
		if moduleName == "" {
			util.PrintColorf(util.LightRed, "[+] Added module name: %s", moduleName)
			modules.Mids.StoreModuleIdentifier(string(inboundRequest.Header.Identifier[:]), inboundRequest.Header.Identifier)
		} else {
			util.PrintColorf(util.LightCyan, "Module name: %s", moduleName)
		}

		// senderModule := modules.GetModule(moduleName)

		source := fmt.Sprint(modules.Mids.GetModuleIdentifier(moduleName)) // Sorry about this (this should print the identifier of the module that sent the request)
		fmt.Println("Source: " + source)

		// m := modules.Modules[string(inboundRequest.Header.Identifier[:])]

		// Process the request...
		// util.PrintColorf(util.BgGreen, "Received: %+v\n", inboundRequest)
		var response []byte

		// Parse metadata
		mData, ok := parseMetadata(d.Metadata)
		if !ok {
			util.PrintWarning("Metadata is nil")
		} else {
			util.PrintSuccess("Metadata: " + fmt.Sprintf("%v", mData))

			// If there is data to fetch, fetch it
			if mData.Method == "GET" {

				//BUG - 🐞: This seems to be caught in an infinite loop. Adding measures to inspect it with user intervention.
				// var message string
				// output := "[🐞] Buggy section ahead! Please read the outputs. press [enter] to continue..."
				// fmt.Scanln(output, &message)

				util.PrintBold("Got a GET request - fetching data...")
				// Get the data source
				// modules.ModuleConfig.DataSources[mData.Destination.Name].GetLogs("path", "filter")
				databaseName := mData.Destination.Object.Database.Name
				tableName := mData.Destination.Object.Database.Table

				// Get the data sources
				util.PrintColorf(util.LightCyan, "Database: %s\nTable: %s", databaseName, tableName)

				// Ask database for the data
				// TODO: Remember to make sure only the latest data is fetched
				// ctx := context.WithValue(*s.moduleStates, "tableName", 0) // TODO: <- Add row id here
				ctx := context.WithValue(context.Background(), tableToModel(tableName), s.moduleStates)
				val := ctx.Value(tableToModel(tableName))

				fmt.Scanln("[⏸️] context value: ", val)

				latestData := fetchLatestLogData(ctx, tableName)

				if latestData.data == nil {
					util.PrintError("Source not found")
					response = []byte("Error, source not found")
					// Respond with an error
					util.PrintError("Failed to fetch the data")

				} else {
					// Get the logs
					response = latestData.data
					// Respond with the data
					util.PrintSuccess("Data fetched successfully")
				}
			}
			if mData.Method == "POST" {
				util.PrintBold("Got a POST request - inserting data...")
				// Insert the data into the database
				// byteData := d["Data"].([]byte)
				// var logdata []ipc.GetJSON

				// json.Unmarshal(byteData, &d)

				// determineActualType(data)

				databaseName := mData.Destination.Object.Database.Name
				tableName := mData.Destination.Object.Database.Table

				// wait := make(chan string)
				// var message string
				bytes, err := json.Marshal(d)
				if err != nil {
					util.PrintError("Failed to marshal the data: " + err.Error())
				}

				fmt.Println("Data: ", string(bytes[:150]))
				// fmt.Printf("Data is : %+v \n", d)
				// m, ok := ipc.GenericData(d["Data"].(map[string]interface{}))
				// if !ok {
				// 	return fmt.Errorf("want type map[string]interface{};  got %T", t)
				// }
				util.PrintDebug("Key value pairs in the data object: ")
				field := 0
				for k, v := range d.Data.(map[string]interface{}) {
					field++
					fmt.Println(k, "=>", v)
				}

				fmt.Println("Field count: ", field)
				// wait <- message

				if d.Data != nil {
					switch d.Data.(type) {
					case []models.AttackType:
						fmt.Println("Data is of type AttackType")
						insertData[[]models.AttackType](databaseName, tableName, d.Data.(map[string]interface{}))
					case []models.NginxLog:
						fmt.Println("Data is of type NginxLog")
						insertData[[]models.NginxLog](databaseName, tableName, d.Data.(map[string]interface{}))
					case []models.SynTraffic:
						fmt.Println("Data is of type SynTraffic")
						insertData[[]models.SynTraffic](databaseName, tableName, d.Data.(map[string]interface{}))
					case []models.GeoData:
						fmt.Println("Data is of type GeoData")
						insertData[[]models.GeoData](databaseName, tableName, d.Data.(map[string]interface{}))
					case []models.GeoLocationData:
						fmt.Println("Data is of type GeoLocationData")
						insertData[[]models.GeoLocationData](databaseName, tableName, d.Data.(map[string]interface{}))
					default:
						fmt.Println("Data is of unknown type")
					}

				} else {
					util.PrintError(" [handleconnection] Data is nil")
				}
			}
		}

		if err := reply(response, c, inboundRequest, s); err != nil {
			util.PrintError("handleConnection: " + err.Error())
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
		util.PrintError("Checksum error")
		return fmt.Errorf(string(response))
	}

	err = s.respond(c, response, moduleId)
	if err != nil {
		util.PrintError("handleConnection: " + err.Error())
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
		default:
			return nil
		}
	}
	return nil
}

// WIP: ✅ Tested - working
func insertData[StoreType any](databaseName, tableName string, data ipc.GenericData) {
	// Get the data store
	d := data["data"].([]interface{})

	s, err := stores.Use(tableName)
	if err != nil {

		s := &database.DataStore[StoreType]{}

		util.PrintError("Failed to get the data store: " + s.Name())
	}

	util.PrintDebug("Getting the data store...")
	// Insert the data into the database

	// Depending on the table name, insert the data into the database
	switch tableName {
	case models.ATTACK_TYPE:
		// Insert the data into the database
		util.PrintDebug("Inserting data into the database...")
		/* ERROR
				[DEBUG] Inserting data into the database...
		panic: interface conversion: interface {} is ipc.GenericData, not models.AttackType

		goroutine 10 [running]:
		github.com/pynezz/bivrost/internal/ipc/ipcserver.insertData({0x9f0cbc?, 0x9dfd26?}, {0xc0015bc2d0, 0xc}, {0x96d6a0, 0xc000430690})
				/mnt/c/Users/kada
		*/

		var tmpData []models.AttackType
		tmpBytes, err := json.Marshal(d)
		if err != nil {
			util.PrintError("Failed to marshal the data: " + err.Error())
		}
		err = json.Unmarshal(tmpBytes, &tmpData)
		if err != nil {
			util.PrintError("Failed to unmarshal the data: " + err.Error())
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
			util.PrintError("Failed to insert the data: " + err.Error())
		} else {
			util.PrintSuccess("Data inserted successfully")
		}

		// err := s.AttackTypeStore.InsertLog(d.(models.AttackType))
		// if err != nil {
		// 	util.PrintError("Failed to insert the data: " + err.Error())
		// }

	case models.NGINX_LOGS:
		return
	case models.SYN_TRAFFIC:
		// err := s.SynTrafficStore.InsertLog(data.(models.SynTraffic))
		// if err != nil {
		// 	util.PrintError("Failed to insert the data: " + err.Error())
		// }
	case models.GEO_DATA:
		// err := s.GeoDataStore.InsertLog(data.(models.GeoData))
		// if err != nil {
		// 	util.PrintError("Failed to insert the data: " + err.Error())
		// }
	case models.GEO_LOCATION_DATA:
		// err := s.GeoLocationDataStore.InsertLog(data.(models.GeoLocationData))
		// if err != nil {
		// 	util.PrintError("Failed to insert the data: " + err.Error())
		// }
	default:
		util.PrintError("Table not found: " + tableName)
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
		util.PrintDebug(" [modulestate] failed to get the module states map")

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
	util.PrintDebug(" [modulestate] fetching the latest log data...")
	s, err := stores.Use(tableName)
	if err != nil {
		util.PrintError(" [modulestate] failed to get the data store: " + s.NginxLogStore.Name())
	}

	// Use a read lock to safely read the LastRowID
	moduleState.RLock()
	lastRowID := moduleState.LastRowID
	moduleState.RUnlock()

	util.PrintDebug("Getting the data store...")

	logs, err := s.NginxLogStore.GetLogRangeFromID(lastRowID) // The store is also an issue here....
	if err != nil {
		util.PrintError("[modulestate] failed to get all logs: " + err.Error())
	}

	util.PrintDebug(" [modulestate] getting all logs starting with " + fmt.Sprintf("%d", lastRowID) + "...")
	fmt.Scanln("Press [Enter] to continue...")

	returnData := LogData{
		lastRowID: lastRowID,
		data:      nil,
	}

	returnData.data, err = json.Marshal(&logs)
	if err != nil {
		util.PrintError("Failed to marshal the logs: " + err.Error())
	}

	moduleState.Lock()
	moduleState.LastRowID += len(logs) // Also a bug was here (+ len(logs) instead of += len(logs))
	moduleState.Unlock()
	util.PrintDebug(" [modulestate] updated the last row ID to " + fmt.Sprintf("%d", moduleState.LastRowID))
	util.PrintDebug(" [modulestate] Returning the data...")
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
// req is the request from the client, but // TODO: It should be the request to the client
func (s *IPCServer) respond(c net.Conn, data []byte, moduleId string) error {
	util.PrintDebug("Responding to the client...")

	var response *ipc.IPCRequest
	var err error
	// TODO: Implement different responses based on verbs/methods

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
	util.PrintColor(util.BgGreen, "🚀 Response sent!")

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
		util.PrintError("NewIPCID(): Identifier length must be 4 bytes")
		util.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	ipc.SetIPCID(id)
}

func Cleanup() {
	util.PrintItalic("\t... IPC server cleanup complete.")
}

func crc(b []byte) uint32 {
	return crc32.ChecksumIEEE(b)
}

// TODO's: 1]
// - check the config for activated modules
// - check if the module is already activated
// - check if connection with the module already exists
// - create socket names based on the module name
