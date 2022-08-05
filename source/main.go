package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	PROJECT_NAME = "RCON"
	PROJECT_VERSION = "1.1.0"

	AUTHOR_NAME = "viral32111"
	AUTHOR_WEBSITE = "https://viral32111.com"
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
	flagCommandInterval := 1 // Seconds

	// Setup the command-line flags
	flag.BoolVar( &flagMinecraft, "minecraft", flagMinecraft, "Use the Minecraft protocol and set default port number to 25575." )
	flag.BoolVar( &flagSourceEngine, "sourceengine", flagSourceEngine, "Use the Source Engine protocol and set default port number to 27015." )
	flag.StringVar( &flagAddress, "address", flagAddress, "The IPv4 address of the remote server, e.g. 192.168.0.5." )
	flag.IntVar( &flagPort, "port", flagPort, "The port number of the remote server, e.g. 27020." )
	flag.StringVar( &flagPassword, "password", flagPassword, "The password for the remote console." )
	flag.IntVar( &flagCommandInterval, "interval", flagCommandInterval, "The time to wait in seconds between sending each command. only applicable when multiple commands are specified." )

	// Set a custom help message
	flag.Usage = func() {
		fmt.Printf( "%s, v%s, by %s (%s).\n", PROJECT_NAME, PROJECT_VERSION, AUTHOR_NAME, AUTHOR_WEBSITE )
		fmt.Printf( "\nUsage: %s [-minecraft | -sourceengine] [-address <ip>] [-port <port>] [-password <password>] [-interval <seconds>] <COMMAND> [<COMMAND>, ...]\n", os.Args[ 0 ] )

		flag.PrintDefaults()

		os.Exit( 1 ) // By default it exits with code 2
	}

	// Parse command-line flags & arguments
	flag.Parse()
	argumentCommands := flag.Args()

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
	if ( flagPort <= 0 || flagPort >= 65536 ) {
		fmt.Fprintln( os.Stderr, "Invalid remote server port number specified, must be between 1 and 65535." )
		os.Exit( 1 )
	}

	// Require a valid interval
	if ( flagCommandInterval <= 0 ) {
		fmt.Fprintln( os.Stderr, "Invalid remote server port number specified, must be more than 0 seconds." )
		os.Exit( 1 )
	}

	// Require a command
	if ( len( argumentCommands ) <= 0 ) {
		fmt.Fprintln( os.Stderr, "No command to execute specified." )
		os.Exit( 1 )
	}

	// Connect to remote server
	remoteConnection, dialError := net.Dial( "tcp4", fmt.Sprintf( "%s:%d", ipAddress, flagPort ) )
	if ( dialError != nil ) {
		fmt.Fprintln( os.Stderr, "Error dialing remote server:", dialError.Error() )
		os.Exit( 1 )
	}
	defer remoteConnection.Close() // Disconnect when finished

	// Try to authenticate
	authSuccesful := attemptAuthentication( remoteConnection, flagPassword, flagSourceEngine )
	if ( !authSuccesful ) {
		fmt.Fprintln( os.Stderr, "Failed to authenticate with remote server (wrong password?)." )
		os.Exit( 1 )
	}

	// Loop through each of the commands
	for index, command := range( argumentCommands ) {
		
		// Send the command and print the server's response
		commandResponse := executeCommand( remoteConnection, command, flagSourceEngine )
		fmt.Println( commandResponse )

		// Wait the specified interval before sending the next command, if there is one
		if ( ( index + 1 ) < len( argumentCommands ) ) {
			time.Sleep( time.Duration( flagCommandInterval ) * time.Second )
		}

	}

	// Exit with success
	os.Exit( 0 )

}
