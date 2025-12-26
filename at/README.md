# AT Command Package

A Go language AT command package for managing modems, providing a concise API to send AT commands and handle modem responses.

## Features

- 🚀 **Asynchronous Processing**: Uses goroutines for asynchronous command execution and response handling
- 🔧 **Modular Design**: Clear file structure with well-separated responsibilities
- 📡 **Command Management**: Supports standard AT commands and SMS commands
- 🔔 **Asynchronous Indications**: Handles modem asynchronous notifications
- ⚡ **Timeout Control**: Configurable command timeout mechanism
- 🛡️ **Connection Management**: Automatic connection status detection and error handling

## File Structure

```text
at/
├── at.go          # Core structs and constructors
├── cmd.go         # Command execution functionality
├── indication.go  # Asynchronous indication handling
├── loop.go        # Internal loop processing
├── options.go     # Configuration options system
├── error.go       # Error type definitions
├── parser.go      # Command parsing helper functions
└── README.md      # This document
```

## Installation

```go
go get github.com/rehiy/modem
```

## Quick Start

```go
package main

import (
    "fmt"
    "io"
    "github.com/rehiy/modem/at"
)

func main() {
    // Assume modem is a device implementing io.ReadWriter interface
    var modem io.ReadWriter
    
    // Create AT instance with initialization commands
    atModem := at.New(modem, at.WithCmds("Z", "E0", "+CMEE=1"))
    
    // Send AT command
    response, err := atModem.Command("AT+CSQ")
    if err != nil {
        fmt.Printf("Command execution error: %v\n", err)
        return
    }
    
    fmt.Printf("Signal strength: %s\n", response)
}
```

## API Documentation

### Core Types

#### AT Struct

```go
type AT struct {
    // Internal fields, accessed through methods
}
```

Main Methods:

- `New(modem io.ReadWriter, options ...Option) *AT` - Create new AT instance
- `Command(cmd string, options ...CommandOption) ([]string, error)` - Execute AT command
- `SMSCommand(cmd string, sms string, options ...CommandOption) ([]string, error)` - Execute SMS-related commands
- `AddIndication(prefix string, handler InfoHandler, options ...IndicationOption) error` - Add asynchronous indication handler
- `CancelIndication(prefix string)` - Remove indication handler
- `Closed() <-chan struct{}` - Get connection status channel

### Configuration Options

#### Constructor Options

- `WithEscTime(d time.Duration) EscTimeOption` - Set escape guard time (default: 20ms)
- `WithCmds(cmds ...string) CmdsOption` - Set initialization commands (default: ATZ, ATE0)
- `WithTimeout(d time.Duration) TimeoutOption` - Set command timeout (default: 1s)

#### Command Options

- `WithTimeout(d time.Duration) TimeoutOption` - Set individual command timeout

#### Indication Options

- `WithTrailingLines(l int) TrailingLinesOption` - Set number of trailing lines for indications
- `WithTrailingLine` - Predefined option for one trailing line

### Error Types

Package defines specific error types for different scenarios:

- `ErrClosed` - Operation cannot be performed as modem has been closed
- `ErrDeadlineExceeded` - Modem failed to complete operation within required time
- `ErrError` - Modem returned generic AT ERROR
- `ErrIndicationExists` - Indication already registered for prefix
- `CMEError` - CME Error returned by modem
- `CMSError` - CMS Error returned by modem
- `ConnectError` - Dial attempt failed

### Type Definitions

- `InfoHandler func([]string)` - Handler function for indication information
- `IndicationOption` - Interface for indication configuration options

### API Reference Table

| Category | Method/Type | Description | Parameters | Returns |
|----------|-------------|-------------|------------|---------|
| **Constructor** | `New` | Create AT instance | `modem io.ReadWriter`, `options ...Option` | `*AT` |
| **Core Methods** | `Command` | Execute AT command | `cmd string`, `options ...CommandOption` | `[]string, error` |
| | `SMSCommand` | Execute SMS command | `cmd string`, `sms string`, `options ...CommandOption` | `[]string, error` |
| | `AddIndication` | Register indication handler | `prefix string`, `handler InfoHandler`, `options ...IndicationOption` | `error` |
| | `CancelIndication` | Remove indication handler | `prefix string` | - |
| | `Closed` | Get connection status | - | `<-chan struct{}` |
| **Options** | `WithEscTime` | Set escape guard time | `d time.Duration` | `EscTimeOption` |
| | `WithCmds` | Set init commands | `cmds ...string` | `CmdsOption` |
| | `WithTimeout` | Set timeout | `d time.Duration` | `TimeoutOption` |
| | `WithTrailingLines` | Set trailing lines | `l int` | `TrailingLinesOption` |
| **Errors** | `ErrClosed` | Connection closed error | - | `error` |
| | `ErrDeadlineExceeded` | Timeout error | - | `error` |
| | `CMEError` | CME error type | - | `error` |
| | `CMSError` | CMS error type | - | `error` |

### Common AT Commands Reference

| Command | Description | Usage Example |
|---------|-------------|---------------|
| `ATI` | Manufacturer identification | `atModem.Command("I")` |
| `AT+CSQ` | Signal quality | `atModem.Command("+CSQ")` |
| `AT+CGMI` | Manufacturer info | `atModem.Command("+CGMI")` |
| `AT+CGMM` | Model info | `atModem.Command("+CGMM")` |
| `AT+CMGF=1` | Set text mode | `atModem.Command("+CMGF=1")` |
| `AT+CMGS` | Send SMS | `atModem.SMSCommand("+CMGS=\"phone\"", "message")` |
| `AT+CNUM` | Own number | `atModem.Command("+CNUM")` |

**Note**: Commands should not include "AT" prefix or line endings - these are automatically handled by the package.

### Best Practices

#### 1. Error Handling

Always check for errors after command execution:

```go
response, err := atModem.Command("+CSQ")
if err != nil {
    // Handle specific error types appropriately
    return
}
```

#### 2. Timeout Configuration

Set appropriate timeouts based on command complexity:

```go
// Quick commands - short timeout
atModem.Command("E0", at.WithTimeout(2*time.Second))

// Network operations - longer timeout  
atModem.Command("+COPS?", at.WithTimeout(30*time.Second))
```

#### 3. Indication Management

Register indications only when needed and cancel when done:

```go
// Register for SMS reception
atModem.AddIndication("+CMT:", smsHandler)

// ... later when SMS handling is no longer needed
atModem.CancelIndication("+CMT:")
```

#### 4. Resource Cleanup

Monitor connection status and handle disconnections:

```go
go func() {
    <-atModem.Closed()
    log.Printf("Modem disconnected - performing cleanup")
    // Cancel all active indications, stop goroutines, etc.
}()
```

### Performance Considerations

- **Concurrency**: All modem access is serialized through channels for thread safety
- **Buffer Sizes**: Internal channels have reasonable defaults (command: 10, lines: 50)
- **Memory Usage**: Lines are processed as they arrive, minimizing memory footprint
- **Timeout Handling**: Proper timeout management prevents blocking operations

### Common Pitfalls

1. **Forgetting Error Checks**: Always check command execution results
2. **Incorrect Command Format**: Don't include "AT" prefix or line endings
3. **Missing Timeouts**: Long-running commands should have appropriate timeouts
4. **Resource Leaks**: Cancel indications when no longer needed
5. **Race Conditions**: Use the provided concurrency-safe API methods

### Testing and Debugging

For debugging, you can enable verbose logging or use shorter timeouts:

```go
// Debug configuration with shorter timeouts
atModem := at.New(modem,
    at.WithTimeout(5*time.Second),  // Shorter timeout for debugging
    at.WithCmds("E1"),              // Enable echo for debugging
)
```

This comprehensive API documentation should help users effectively utilize the AT command package for modem communication tasks.

### Detailed API Usage

#### Command Method

```go
// Execute standard AT command
response, err := atModem.Command("AT+CSQ")
if err != nil {
    log.Printf("Signal query failed: %v", err)
} else {
    log.Printf("Signal quality response: %v", response)
}

// Execute command with timeout option
response, err := atModem.Command("AT+CGMI", at.WithTimeout(30*time.Second))
if err != nil {
    log.Printf("Manufacturer query failed: %v", err)
} else {
    log.Printf("Manufacturer: %v", response)
}
```

#### SMSCommand Method

```go
// Execute SMS command in text mode
response, err := atModem.SMSCommand("+CMGS=\"+1234567890\"", "Hello World!")
if err != nil {
    log.Printf("SMS send failed: %v", err)
} else {
    log.Printf("SMS sent successfully: %v", response)
}

// Execute SMS command with timeout
response, err := atModem.SMSCommand("+CMGS=\"+1234567890\"", "Test message", 
    at.WithTimeout(60*time.Second))
```

#### Indication Handling

```go
// Define indication handler
smsHandler := func(lines []string) {
    log.Printf("Received SMS indication: %v", lines)
    // Process SMS content from lines
}

// Register SMS reception indication
err := atModem.AddIndication("+CMT:", smsHandler, at.WithTrailingLines(1))
if err != nil {
    log.Printf("Indication registration failed: %v", err)
    return
}

// Cancel indication when no longer needed
atModem.CancelIndication("+CMT:")
```

#### Connection Status Monitoring

```go
// Monitor connection status
go func() {
    select {
    case <-atModem.Closed():
        log.Printf("Modem connection closed")
        // Perform cleanup operations
    }
}()
```

#### Error Handling Examples

```go
// Handle specific error types
response, err := atModem.Command("AT+INVALID")
if err != nil {
    switch err {
    case at.ErrDeadlineExceeded:
        log.Printf("Command timeout - modem not responding")
    case at.ErrClosed:
        log.Printf("Connection closed - modem disconnected")
    case at.ErrError:
        log.Printf("Modem returned ERROR - invalid command")
    case at.CMEError:
        log.Printf("CME Error: %v", err)
    case at.CMSError:
        log.Printf("CMS Error: %v", err)
    default:
        log.Printf("Command execution error: %v", err)
    }
}
```

### Advanced Usage

#### Custom Initialization Commands

```go
// Create AT instance with custom initialization
atModem := at.New(modem,
    at.WithCmds("Z", "E0", "+CMEE=1"),  // Reset, disable echo, enable verbose errors
    at.WithEscTime(50*time.Millisecond),  // Set longer escape time
    at.WithTimeout(5*time.Second),         // Set default command timeout
)
```

#### Multiple Indication Handlers

```go
// Register multiple indications
callHandler := func(lines []string) {
    log.Printf("Incoming call: %v", lines)
}

networkHandler := func(lines []string) {
    log.Printf("Network status change: %v", lines)
}

atModem.AddIndication("RING", callHandler)
atModem.AddIndication("+CREG:", networkHandler, at.WithTrailingLine)
```

### Parameter Notes

- **Command strings**: Should not include "AT" prefix or "\r\n" suffix (automatically added)
- **SMS commands**: Two-step process - command line followed by SMS data
- **Timeout values**: Use `time.Duration` (e.g., `30*time.Second`)
- **Indication prefixes**: Match modem response prefixes exactly
- **Handler functions**: Receive slice of strings containing indication lines

### Design Philosophy

#### Concurrency Safety

All modem access is serialized through channels, ensuring concurrency safety.

#### Module Separation

- **at.go**: Core coordination and lifecycle management
- **cmd.go**: Command execution and response handling
- **indication.go**: Asynchronous notification handling
- **parser.go**: Response parsing logic
- **options.go**: Configuration options system

#### Error Handling

Provides clear error types and detailed error information for easy debugging and troubleshooting.

### Contribution Guidelines

Welcome to submit Issues and Pull Requests to improve this package.

### License

[MIT License](LICENSE)

---

## Acknowledgments

This project is based on the AT command package implementation from [warthog618/modem](https://github.com/warthog618/modem/blob/master/at/at.go).

Special thanks to the original author for the excellent design and implementation, which provided us with a stable and reliable foundation for AT command processing. We have performed modular refactoring and functional enhancements on this basis, but the core design philosophy and architectural ideas all originate from the original project.

**Salute to the original author's open source contribution!** 🙏
