# bluelock-go

A Go application for data pulling and integration services with a flexible dependency management pattern.

## Architecture Overview

This project implements a **Hybrid Singleton Pattern** that combines the benefits of singletons for production use with direct constructors for testing, providing a pragmatic solution for dependency management in Go applications.

## Dependency Management Pattern

### The Hybrid Singleton Pattern

Our codebase uses a sophisticated pattern that addresses the common trade-offs between simplicity and testability in Go applications.

#### Pattern Components

**1. Production Singleton Behavior**
```go
// Production code uses Acquire*() methods for easy access
func AcquireBitbucketCloudSvc() *BitbucketCloudSvc {
    if bitbucketCloudSvc == nil {
        // Lazy initialization with dependency acquisition
        customLogger := shared.AcquireCustomLogger()
        cfg := config.AcquireConfig()
        statemanager := statemanager.AcquireStateManager()
        credentials := credservice.AcquireCredentials()
        dbQuerier := dbsetup.AcquireQueries()
        client := AcquireClient()
        
        bitbucketCloudSvc = NewBitbucketCloudSvc(customLogger, statemanager, 
                                                credentials, cfg, dbQuerier, client)
    }
    return bitbucketCloudSvc
}
```

**2. Testing Direct Constructor**
```go
// Test code uses direct constructors with mocked dependencies
func TestRepoPull() {
    mockLogger := &MockLogger{}
    mockStateManager := &MockStateManager{}
    mockCredentials := []auth.Credential{...}
    mockConfig := &config.Config{...}
    mockDBQuerier := &MockDBQuerier{}
    mockClient := &MockClient{}
    
    service := NewBitbucketCloudSvc(mockLogger, mockStateManager, 
                                   mockCredentials, mockConfig, mockDBQuerier, mockClient)
    err := service.RepoPull()
    // Pure unit test with controlled dependencies
}
```

#### Pattern Benefits

**Advantages over Pure Singletons:**
- ✅ **Testable**: Can use mocks in tests without global state pollution
- ✅ **Flexible**: Can override dependencies when needed
- ✅ **Clear Dependencies**: Constructor parameters explicitly show what's needed
- ✅ **No Race Conditions**: Lazy initialization with proper synchronization

**Advantages over Full DI Frameworks:**
- ✅ **Simpler**: No complex wiring or interface definitions
- ✅ **Familiar**: Uses Go patterns developers already know
- ✅ **Lightweight**: No external dependencies or reflection overhead
- ✅ **Fast**: Direct function calls, no dependency resolution

**Advantages over Manual DI:**
- ✅ **Production Ready**: Singleton behavior for production simplicity
- ✅ **Less Boilerplate**: No need to wire everything in main functions
- ✅ **Consistent**: Same pattern across all services

#### Implementation Details

**Service Structure**
```go
type BitbucketCloudSvc struct {
    logger       *shared.CustomLogger
    stateManager *statemanager.StateManager
    credentials  []auth.Credential
    config       *config.Config
    apiClient    *Client
    dbQuerier    dbgen.Querier
}

// Two constructors for different use cases
func NewBitbucketCloudSvc(logger, stateManager, credentials, config, dbQuerier, client) *BitbucketCloudSvc {
    // Direct constructor for testing and explicit dependency injection
    return &BitbucketCloudSvc{logger, stateManager, credentials, config, client, dbQuerier}
}

func AcquireBitbucketCloudSvc() *BitbucketCloudSvc {
    // Singleton constructor for production use
    // Uses global state for dependency acquisition
}
```

**Main Function Pattern**
```go
func main() {
    // Initialize dependencies in specific order
    // 1. Logger (everything depends on it)
    logFile, err := shared.InitializeCustomLogger(appLoggerFilePath, shared.TextLogHandler)
    customLogger := shared.AcquireCustomLogger()
    
    // 2. Credentials (needed by state manager)
    if err = credservice.InitializeAuthCredentialStore(authTokensFilePath, credservice.DatapullCredentialsKey); err != nil {
        // handle error
    }
    
    // 3. State Manager (depends on credentials)
    if err := statemanager.InitializeStateManager(stateJsonFilePath); err != nil {
        // handle error
    }
    
    // 4. Config (needed by services)
    if err := config.InitializeConfig(); err != nil {
        // handle error
    }
    
    // 5. Database (depends on config)
    db, err := dbsetup.InitializeDb()
    defer db.Close()
    
    // Use singleton pattern for services
    service := bitbucketcloud.AcquireBitbucketCloudSvc()
    service.RunJob()
}
```

**Test Function Pattern**
```go
func TestService() {
    // Use direct constructors with mocks
    service := NewBitbucketCloudSvc(mockLogger, mockStateManager, ...)
    result := service.RunJob()
    // Assert on result
}
```

#### Initialization Order

The pattern requires careful initialization order in main functions:

1. **Logger** - Everything depends on logging
2. **Credentials** - Needed by state manager
3. **State Manager** - Depends on credentials
4. **Config** - Needed by services
5. **Database** - Depends on config
6. **Integration Services** - Depend on everything above

**Important**: This order must be maintained to prevent initialization errors.

#### When to Use This Pattern

✅ **Use this pattern when:**
- You want simplicity in production code
- You need good testability without complex DI frameworks
- You're building data processing applications
- You have a small to medium team
- You want to avoid external dependencies
- You prefer Go idioms over framework abstractions

❌ **Consider alternatives when:**
- You have very complex dependency graphs
- You need dynamic service resolution
- You're building large microservices architectures
- You have many different service types with varying dependencies
- You need advanced DI features like lifecycle management

#### Pattern Evolution Path

This pattern can evolve as your application grows:

1. **Phase 1**: Current hybrid pattern (current state)
2. **Phase 2**: Add dependency validation and health checks
3. **Phase 3**: Add context support for better control
4. **Phase 4**: Consider full DI only if complexity demands it

#### Best Practices

1. **Always provide both constructors**: `New*()` for testing, `Acquire*()` for production
2. **Document initialization order**: Clearly specify dependency order in main functions
3. **Use mocks in tests**: Leverage the direct constructor pattern for isolated unit tests
4. **Keep services focused**: Each service should have clear, minimal dependencies
5. **Validate dependencies**: Add validation to prevent runtime errors
6. **Use consistent naming**: Follow the `New*()` and `Acquire*()` naming convention

#### Example Service Implementation

```go
// integrations/git/bitbucket/bitbucketcloud/mainsvc.go
package bitbucketcloud

type BitbucketCloudSvc struct {
    logger       *shared.CustomLogger
    stateManager *statemanager.StateManager
    credentials  []auth.Credential
    config       *config.Config
    apiClient    *Client
    dbQuerier    dbgen.Querier
}

// Direct constructor for testing and explicit DI
func NewBitbucketCloudSvc(logger *shared.CustomLogger, stateManager *statemanager.StateManager, 
                         credentials []auth.Credential, config *config.Config, 
                         dbQuerier dbgen.Querier, client *Client) *BitbucketCloudSvc {
    return &BitbucketCloudSvc{logger, stateManager, credentials, config, client, dbQuerier}
}

// Singleton constructor for production use
func AcquireBitbucketCloudSvc() *BitbucketCloudSvc {
    if bitbucketCloudSvc == nil {
        customLogger := shared.AcquireCustomLogger()
        cfg := config.AcquireConfig()
        statemanager := statemanager.AcquireStateManager()
        credentials := credservice.AcquireCredentials()
        dbQuerier := dbsetup.AcquireQueries()
        client := AcquireClient()
        
        bitbucketCloudSvc = NewBitbucketCloudSvc(customLogger, statemanager, 
                                                credentials, cfg, dbQuerier, client)
    }
    return bitbucketCloudSvc
}
```

## Project Structure

```
bluelock-go/
├── cmd/
│   ├── datapuller/     # Main data pulling application
│   └── authsync/       # Authentication synchronization
├── config/             # Configuration management
├── integrations/       # Integration services
│   └── git/
│       └── bitbucket/
│           └── bitbucketcloud/
├── shared/             # Shared utilities and services
│   ├── auth/           # Authentication
│   ├── database/       # Database operations
│   ├── storage/        # State management
│   └── jobscheduler/   # Job scheduling
└── secrets/            # Configuration files
```

## Getting Started

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd bluelock-go
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up configuration**
   ```bash
   cp config/config.json config/config.user.json
   # Edit config.user.json with your settings
   ```

4. **Build the application**
   ```bash
   make build
   ```

5. **Run the data puller**
   ```bash
   make run-puller
   ```

## Testing

The hybrid singleton pattern makes testing straightforward:

```bash
# Run all tests
make test

# Run specific package tests
go test ./integrations/git/bitbucket/bitbucketcloud
```

## Contributing

When adding new services, follow the hybrid singleton pattern:

1. Create a service struct with clear dependencies
2. Implement both `New*()` and `Acquire*()` constructors
3. Add the service to the appropriate initialization order in main
4. Write tests using the direct constructor with mocks
