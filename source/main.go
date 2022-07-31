// go build -o rcon.exe ./source/
// go run ./source/

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

func main() {

	// Seed the random number generator
	rand.Seed( time.Now().UnixNano() )

	// Variables for the command-line flags with their default values
	flagMinecraft := false
	flagSourceEngine := false
	flagAddress := "127.0.0.1"
	flagPort := 0
	flagPassword := ""

	// Setup the command-line flags
	flag.BoolVar( &flagMinecraft, "minecraft", flagMinecraft, "Use the Minecraft protocol and set default port number to 25575." )
	flag.BoolVar( &flagSourceEngine, "sourceengine", flagSourceEngine, "Use the Source Engine protocol and set default port number to 27015." )
	flag.StringVar( &flagAddress, "address", flagAddress, "The IPv4 address of the remote server." )
	flag.IntVar( &flagPort, "port", flagPort, "The port number of the remote server." )
	flag.StringVar( &flagPassword, "password", flagPassword, "The password for the remote console." )

	// Parse command-line flags & arguments
	flag.Parse()
	argCommand := strings.Join( flag.Args(), " " )

	// Require either the Minecraft or the Source Engine flag, but never neither nor both
	if ( flagMinecraft && flagSourceEngine ) {
		fmt.Fprintln( os.Stderr, "The -minecraft and -sourceengine flags cannot be used together." )
		os.Exit( 1 )
	} else if ( !flagMinecraft && !flagSourceEngine ) {
		fmt.Fprintln( os.Stderr, "Either the -minecraft or the -sourceengine flag must be specified." )
		os.Exit( 1 )
	}

	// Set the port number depending on the protocol, but only if a custom port was not specified
	if ( flagPort == 0 ) {
		if ( flagMinecraft ) {
			flagPort = 25575
		} else if ( flagSourceEngine ) {
			flagPort = 27015
		}
	}

	// Require a valid IP address
	ipAddress := net.ParseIP( flagAddress )
	if ( flagAddress == "" || ipAddress == nil || ipAddress.To4() == nil ) {
		fmt.Fprintln( os.Stderr, "Invalid remote server IPv4 address specified." )
		os.Exit( 1 )
	}

	// Require a valid port number
	if ( flagPort < 0 || flagPort > 65536 ) {
		fmt.Fprintln( os.Stderr, "Invalid remote server port number specified." )
		os.Exit( 1 )
	}

	// Require a command
	if ( argCommand == "" ) {
		fmt.Fprintln( os.Stderr, "Invalid command to execute specified." )
		os.Exit( 1 )
	}

	// Connect to remote server
	remoteConnection, dialError := net.Dial( "tcp4", fmt.Sprintf( "%s:%d", ipAddress, flagPort ) )
	if ( dialError != nil ) {
		fmt.Fprintln( os.Stderr, "Error dialing remote server:", dialError.Error() )
		os.Exit( 1 )
	}
	defer remoteConnection.Close() // Disconnect when finished

	authSuccesful := attemptAuthentication( remoteConnection, flagPassword, flagSourceEngine )
	if ( !authSuccesful ) {
		fmt.Fprintln( os.Stderr, "Failed to authenticate with remote server (wrong password?)." )
		os.Exit( 1 )
	}

	commandResponse := executeCommand( remoteConnection, argCommand, flagSourceEngine )
	fmt.Println( commandResponse )

}
