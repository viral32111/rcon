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

/*
https://developer.valvesoftware.com/wiki/Source_RCON_Protocol
https://wiki.vg/RCON

SIZE: int32 LE [4 bytes] - Total byte length of ID + TYPE + BODY + NT
ID: int32 LE [4 bytes] - Unique, chosen by client for each request, must be positive integer
TYPE: int32 LE [4 bytes] - The type of packet (0, 2 or 3)
BODY: null terminated ASCII string) [n+1 bytes] - Password for auth type, or command to execute
NULL TERMINATOR (0x00) [1 bytes]
*/

const MAX_RESPONSE_PACKET_SIZE = 4096

const TYPE_AUTH = 3 // LOGIN, SERVERDATA_AUTH
const TYPE_AUTH_RESPONSE = 2 // SERVERDATA_AUTH_RESPONSE
const TYPE_COMMAND = 2 // COMMAND, SERVERDATA_EXECCOMMAND
const TYPE_COMMAND_RESPONSE = 0 // COMMAND RESPONSE, SERVERDATA_RESPONSE_VALUE

type ResponsePacket struct {
	ID uint32
	Type uint32
	Body string
}

func attemptAuthentication( connection net.Conn, password string, isSourceEngine bool ) bool {

	// Send authentication request
	authRequestPacket := createPacket( TYPE_AUTH, password )
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

	// Receive empty command response if this is for the Source Enginne
	if ( isSourceEngine ) {
		commandResponseData := make( []byte, MAX_RESPONSE_PACKET_SIZE )
		commandResponseBytesRead, commandResponseReadError := connection.Read( commandResponseData )
		if ( commandResponseReadError != nil ) {
			fmt.Fprintln( os.Stderr, "Error reading response value from remote connection:", commandResponseReadError.Error() )
			os.Exit( 1 )
		}
		//fmt.Printf( "Read %d bytes: %s\n", commandResponseBytesRead, hex.EncodeToString( commandResponseData[ : commandResponseBytesRead ] ) )

		commandResponsePacket := parsePacket( commandResponseData[ 0 : commandResponseBytesRead ] )
		if ( commandResponsePacket.Type != TYPE_COMMAND_RESPONSE ) {
			fmt.Fprintln( os.Stderr, "Received unexpected packet type (expecting command response):", commandResponsePacket.Type )
			os.Exit( 1 )
		}
	}

	// Receive authentication response
	authResponseData := make( []byte, MAX_RESPONSE_PACKET_SIZE )
	authResponseBytesRead, authResponseError := connection.Read( authResponseData )
	if ( authResponseError != nil ) {
		fmt.Fprintln( os.Stderr, "Error reading authentication response from remote connection:", authResponseError.Error() )
		os.Exit( 1 )
	}
	//fmt.Printf( "Read %d bytes: %s\n", authResponseBytesRead, hex.EncodeToString( authResponseData[ : authResponseBytesRead ] ) )

	authResponsePacket := parsePacket( authResponseData[ 0 : authResponseBytesRead ] )
	if ( authResponsePacket.Type != TYPE_AUTH_RESPONSE ) {
		fmt.Fprintln( os.Stderr, "Received unexpected packet type (expecting authentication response):", authResponsePacket.Type )
		os.Exit( 1 )
	}

	// Packet ID is -1 if authentication failed
	return authResponsePacket.ID != 0xFFFFFFFF

}

func executeCommand( connection net.Conn, command string, isSourceEngine bool ) string {

	// Send command request
	commandRequestPacket := createPacket( TYPE_COMMAND, command )
	commandRequestBytesWrote, commandRequestWriteError := connection.Write( commandRequestPacket )
	if ( commandRequestWriteError != nil ) {
		fmt.Fprintln( os.Stderr, "Error writing execute command request to remote connection:", commandRequestWriteError.Error() )
		os.Exit( 1 )
	}
	if ( commandRequestBytesWrote != len( commandRequestPacket ) ) {
		fmt.Fprintln( os.Stderr, "Bytes written length mismatch with execute command request packet." )
		os.Exit( 1 )
	}
	//fmt.Printf( "\nWrote %d bytes: %s\n", commandRequestBytesWrote, hex.EncodeToString( commandRequestPacket ) )

	// Receive command response
	commandResponseData := make( []byte, MAX_RESPONSE_PACKET_SIZE )
	commandResponseBytesRead, commandResponseReadError := connection.Read( commandResponseData )
	if ( commandResponseReadError != nil ) {
		fmt.Fprintln( os.Stderr, "Error reading response value from remote connection:", commandResponseReadError.Error() )
		os.Exit( 1 )
	}
	//fmt.Printf( "Read %d bytes: %s\n", commandResponseBytesRead, hex.EncodeToString( commandResponseData[ : commandResponseBytesRead ] ) )

	responseValuePacket := parsePacket( commandResponseData[ 0 : commandResponseBytesRead ] )
	if ( responseValuePacket.Type != TYPE_COMMAND_RESPONSE ) {
		fmt.Fprintln( os.Stderr, "Received unexpected packet type (expecting command response):", responseValuePacket.Type )
		os.Exit( 1 )
	}

	// Remove trailing new line & "rcon from xxx.xxx.xxx.xxx: command xxxxxxxx" line if this is for the Source Engine
	if ( isSourceEngine ) {
		trimmedResponse := strings.TrimRight( responseValuePacket.Body, "\n" )

		finalLineStartPosition := strings.LastIndex( trimmedResponse, "\n" )
		if ( finalLineStartPosition == -1 ) {
			fmt.Fprintln( os.Stderr, "Final new-line not found in command response." )
			os.Exit( 1 )
		}

		commandResponse := trimmedResponse[ 0 : finalLineStartPosition ]

		return commandResponse
	}

	return responseValuePacket.Body

}

func createPacket( requestType int, requestBody string ) []byte {

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

func parsePacket( responsePacket []byte ) ResponsePacket {

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

	return ResponsePacket {
		ID: binary.LittleEndian.Uint32( packetIdentifier ),
		Type: binary.LittleEndian.Uint32( packetType ),
		Body: string( packetBody ),
	}

}
