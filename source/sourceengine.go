package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
)

// https://developer.valvesoftware.com/wiki/Source_RCON_Protocol

// Size: int32 LE [4B] - Total byte length of ID + Type + Body + NT
// ID: int32 LE [4B] - Unique, chosen by client for each request, must be positive integer
// Type: int32 LE [4B] - AUTH (3), AUTH_RESPONSE (2), EXECCOMMAND (2), RESPONSE_VALUE (0)
// Body: ASCII String (null terminated/0x00) [>1B]
// Null Terminator/0x00 [1B]

const SOURCEENGINE_MAX_PACKET_SIZE = 4096

const SOURCEENGINE_SERVERDATA_AUTH = 3
const SOURCEENGINE_SERVERDATA_AUTH_RESPONSE = 2
const SOURCEENGINE_SERVERDATA_EXECCOMMAND = 2
const SOURCEENGINE_SERVERDATA_RESPONSE_VALUE = 0

type SourceEnginePacket struct {
	ID uint32 // Unique (4 bytes)
	Type uint32 // 0, 2, 3 (4 bytes)
	Body string // n + 1 bytes
}

func sourceEngineAuthenticate( connection net.Conn, password string ) bool {

	/************ SERVERDATA_AUTH ************/

	authRequestPacket := sourceEngineCreatePacket( SOURCEENGINE_SERVERDATA_AUTH, password )
	authRequestBytesWrote, authRequestWriteError := connection.Write( authRequestPacket )
	if ( authRequestWriteError != nil ) {
		fmt.Fprintln( os.Stderr, "Error writing authentication request to remote connection:", authRequestWriteError.Error() )
		os.Exit( 1 )
	}
	if ( authRequestBytesWrote != len( authRequestPacket ) ) {
		fmt.Fprintln( os.Stderr, "Bytes written length mismatch with authentication request packet." )
		os.Exit( 1 )
	}
	//fmt.Printf( "\nWrote %d bytes: %s\n", authRequestBytesWrote, hex.EncodeToString( authRequestPacket ) )

	/************ SERVERDATA_RESPONSE_VALUE ************/
	
	responseValueData := make( []byte, SOURCEENGINE_MAX_PACKET_SIZE )
	responseValueBytesRead, responseValueReadError := connection.Read( responseValueData )
	if ( responseValueReadError != nil ) {
		fmt.Fprintln( os.Stderr, "Error reading response value from remote connection:", responseValueReadError.Error() )
		os.Exit( 1 )
	}
	//fmt.Printf( "Read %d bytes: %s\n", responseValueBytesRead, hex.EncodeToString( responseValueData[ : responseValueBytesRead ] ) )

	responseValuePacket := sourceEngineParsePacket( responseValueData[ 0 : responseValueBytesRead ] )
	if ( responseValuePacket.Type != SOURCEENGINE_SERVERDATA_RESPONSE_VALUE ) {
		fmt.Fprintln( os.Stderr, "Received unexpected packet type (expecting SERVERDATA_RESPONSE_VALUE):", responseValuePacket.Type )
		os.Exit( 1 )
	}

	/************ SERVERDATA_AUTH_RESPONSE ************/
	
	authResponseData := make( []byte, SOURCEENGINE_MAX_PACKET_SIZE )
	authResponseBytesRead, authResponseError := connection.Read( authResponseData )
	if ( authResponseError != nil ) {
		fmt.Fprintln( os.Stderr, "Error reading authentication response from remote connection:", authResponseError.Error() )
		os.Exit( 1 )
	}
	//fmt.Printf( "Read %d bytes: %s\n", authResponseBytesRead, hex.EncodeToString( authResponseData[ : authResponseBytesRead ] ) )

	authResponsePacket := sourceEngineParsePacket( authResponseData[ 0 : authResponseBytesRead ] )
	if ( authResponsePacket.Type != SOURCEENGINE_SERVERDATA_AUTH_RESPONSE ) {
		fmt.Fprintln( os.Stderr, "Received unexpected packet type (expecting SERVERDATA_AUTH_RESPONSE):", authResponsePacket.Type )
		os.Exit( 1 )
	}

	return authResponsePacket.ID != 0xFFFFFFFF

}

func sourceEngineExecuteCommand( connection net.Conn, command string ) string {

	/************ SERVERDATA_EXECCOMMAND ************/

	execRequestPacket := sourceEngineCreatePacket( SOURCEENGINE_SERVERDATA_EXECCOMMAND, command )
	execRequestBytesWrote, execRequestWriteError := connection.Write( execRequestPacket )
	if ( execRequestWriteError != nil ) {
		fmt.Fprintln( os.Stderr, "Error writing execute command request to remote connection:", execRequestWriteError.Error() )
		os.Exit( 1 )
	}
	if ( execRequestBytesWrote != len( execRequestPacket ) ) {
		fmt.Fprintln( os.Stderr, "Bytes written length mismatch with execute command request packet." )
		os.Exit( 1 )
	}
	//fmt.Printf( "\nWrote %d bytes: %s\n", execRequestBytesWrote, hex.EncodeToString( execRequestPacket ) )

	/************ SERVERDATA_RESPONSE_VALUE ************/
	
	responseValueData := make( []byte, SOURCEENGINE_MAX_PACKET_SIZE )
	responseValueBytesRead, responseValueReadError := connection.Read( responseValueData )
	if ( responseValueReadError != nil ) {
		fmt.Fprintln( os.Stderr, "Error reading response value from remote connection:", responseValueReadError.Error() )
		os.Exit( 1 )
	}
	//fmt.Printf( "Read %d bytes: %s\n", responseValueBytesRead, hex.EncodeToString( responseValueData[ : responseValueBytesRead ] ) )

	responseValuePacket := sourceEngineParsePacket( responseValueData[ 0 : responseValueBytesRead ] )
	if ( responseValuePacket.Type != SOURCEENGINE_SERVERDATA_RESPONSE_VALUE ) {
		fmt.Fprintln( os.Stderr, "Received unexpected packet type (expecting SERVERDATA_RESPONSE_VALUE):", responseValuePacket.Type )
		os.Exit( 1 )
	}

	// Remove trailing new line & "rcon from xxx.xxx.xxx.xxx: command xxxxxxxx" line
	trimmedResponse := strings.TrimRight( responseValuePacket.Body, "\n" )
	finalLineStartPosition := strings.LastIndex( trimmedResponse, "\n" )
	if ( finalLineStartPosition == -1 ) {
		fmt.Fprintln( os.Stderr, "Final new-line not found in command response." )
		os.Exit( 1 )
	}
	commandResponse := trimmedResponse[ 0 : finalLineStartPosition ]

	return commandResponse

}

func sourceEngineCreatePacket( requestType int, requestBody string ) []byte {

	// ID
	packetIdentifier := make( []byte, 4 )
	binary.LittleEndian.PutUint32( packetIdentifier, rand.Uint32() )
	//fmt.Printf( "\nIdentifier: %d (%d bytes)\n", binary.LittleEndian.Uint32( packetIdentifier ), len( packetIdentifier ) )

	// Type
	packetType := make( []byte, 4 )
	binary.LittleEndian.PutUint32( packetType, uint32( requestType ) )
	//fmt.Printf( "Type: %d (%d bytes)\n", binary.LittleEndian.Uint32( packetType ), len( packetType ) )

	// Body
	packetBody := bytes.NewBuffer( nil )
	packetBody.WriteString( requestBody )
	packetBody.WriteByte( 0x00 )
	//fmt.Printf( "Body: %s (%d bytes)\n", hex.EncodeToString( packetBody.Bytes() ), packetBody.Len() )

	/**********************************/

	// Length of: ID + Type + Body + NT
	packetSize := make( []byte, 4 )
	binary.LittleEndian.PutUint32( packetSize, uint32( 4 + 4 + packetBody.Len() + 1 ) )
	//fmt.Printf( "Size: %d (%d bytes)\n", binary.LittleEndian.Uint32( packetSize ), len( packetSize ) )

	// Everything + NT
	packet := bytes.NewBuffer( nil )
	packet.Write( packetSize ) // 4 bytes
	packet.Write( packetIdentifier ) // 4 bytes
	packet.Write( packetType ) // 4 bytes
	packet.Write( packetBody.Bytes() ) // n + 1 bytes
	packet.WriteByte( 0x00 ) // 1 byte (NT)

	/**********************************/

	return packet.Bytes()

}

func sourceEngineParsePacket( responsePacket []byte ) SourceEnginePacket {

	responseReader := bytes.NewReader( responsePacket )

	/**********************************/

	// Size
	packetSize := make( []byte, 4 )
	responseReader.Read( packetSize )
	//fmt.Printf( "\nSize: %d (%d bytes)\n", binary.LittleEndian.Uint32( packetSize ), len( packetSize ) )

	/**********************************/

	// Identifier
	packetIdentifier := make( []byte, 4 )
	responseReader.Read( packetIdentifier )
	//fmt.Printf( "Identifier: %d (%d bytes)\n", binary.LittleEndian.Uint32( packetIdentifier ), len( packetIdentifier ) )

	// Type
	packetType := make( []byte, 4 )
	responseReader.Read( packetType )
	//fmt.Printf( "Type: %d (%d bytes)\n", binary.LittleEndian.Uint32( packetType ), len( packetType ) )

	// Body
	packetBody := make( []byte, binary.LittleEndian.Uint32( packetSize ) - 4 - 4 - 1 )
	responseReader.Read( packetBody )
	packetBody = packetBody[ 0 : len( packetBody ) - 1 ] // Remove NT
	//fmt.Printf( "Body: %s (%d bytes)\n", hex.EncodeToString( packetBody ), len( packetBody ) )

	/**********************************/

	// NT
	packetNT := make( []byte, 1 )
	responseReader.Read( packetNT )
	//fmt.Printf( "NT: %s (%d bytes)\n\n", hex.EncodeToString( packetNT ), len( packetNT ) )

	/**********************************/

	return SourceEnginePacket {
		ID: binary.LittleEndian.Uint32( packetIdentifier ),
		Type: binary.LittleEndian.Uint32( packetType ),
		Body: string( packetBody ),
	}

}
