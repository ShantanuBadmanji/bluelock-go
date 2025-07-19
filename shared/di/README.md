# Improved Dependency Injection with Interface Segregation

This package provides a sophisticated dependency injection solution that addresses the problems of high-scope access and main-specific containers.

## Problems Solved

### 1. High-Scope Container Access
**Before**: Services had access to dependencies they didn't need
```go
// Client had access to config, db, etc. that it didn't need
func NewClient(container *di.Container) *Client {
    // container.Config, container.DB - not needed!
}
```

**After**: Services only get the dependencies they need
```go
// Client only gets API-related dependencies
func NewClient(apiProvider di.APIProvider) *Client {
    // Only logger, stateManager, credentials, httpClient
}
```

### 2. Main-Specific Containers
**Before**: Each main file needed the same full container
```go
// datapuller/main.go and authsync/main.go both needed full container
container := di.NewContainer().SetLogger(...).SetConfig(...).SetDB(...).SetHTTPClient(...)
```

**After**: Each main file uses only what it needs
```go
// datapuller/main.go - needs everything
container := di.NewContainer().SetLogger(...).SetConfig(...).SetDB(...).SetHTTPClient(...)

// authsync/main.go - needs only basic dependencies
serviceContainer := di.NewServiceContainer().SetLogger(...).SetConfig(...).SetCredentials(...)
```

## Architecture

### Interface Segregation
We define specific interfaces for different dependency needs:

```go
// Basic providers
type LoggerProvider interface { GetLogger() *shared.CustomLogger }
type ConfigProvider interface { GetConfig() *config.Config }
type StateManagerProvider interface { GetStateManager() *statemanager.StateManager }
type CredentialsProvider interface { GetCredentials() []auth.Credential }
type HTTPClientProvider interface { GetHTTPClient() *http.Client }
type DatabaseProvider interface { GetDB() *sql.DB; GetDBQueries() *dbgen.Queries }

// Combined providers for specific use cases
type APIProvider interface {
    LoggerProvider
    StateManagerProvider
    CredentialsProvider
    HTTPClientProvider
}

type ServiceProvider interface {
    LoggerProvider
    ConfigProvider
    StateManagerProvider
    CredentialsProvider
}

type FullProvider interface {
    LoggerProvider
    ConfigProvider
    StateManagerProvider
    CredentialsProvider
    HTTPClientProvider
    DatabaseProvider
}
```

### Specialized Containers
Different container types for different needs:

```go
// APIContainer - for API services (Client, etc.)
type APIContainer struct {
    Logger       *shared.CustomLogger
    StateManager *statemanager.StateManager
    Credentials  []auth.Credential
    HTTPClient   *http.Client
}

// ServiceContainer - for main services (BitbucketCloudSvc, etc.)
type ServiceContainer struct {
    Logger       *shared.CustomLogger
    Config       *config.Config
    StateManager *statemanager.StateManager
    Credentials  []auth.Credential
}

// DatabaseContainer - for database operations
type DatabaseContainer struct {
    DB        *sql.DB
    DBQueries *dbgen.Queries
}
```

### Flexible Validation
Different validation strategies for different use cases:

```go
// Full container validation (for datapuller that needs everything)
container := di.NewContainer().SetLogger(...).SetConfig(...).SetDB(...).SetHTTPClient(...)
if err := container.Validate(); err != nil {
    // Validates ALL dependencies are present
}

// Service-only validation (for authsync that doesn't need DB/HTTP)
serviceContainer := container.ToServiceContainer()
if err := serviceContainer.Validate(); err != nil {
    // Validates only service dependencies (no DB, no HTTP client)
}

// API-only validation (for API services)
apiContainer := container.ToAPIContainer()
if err := apiContainer.Validate(); err != nil {
    // Validates only API dependencies (no DB, no config)
}

// Database-only validation
databaseContainer := container.ToDatabaseContainer()
if err := databaseContainer.Validate(); err != nil {
    // Validates only database dependencies
}
```

### Interface Implementation Checks
Compile-time safety for interface compliance:

```go
// In specialized_containers.go
var _ APIProvider = (*APIContainer)(nil)
var _ ServiceProvider = (*ServiceContainer)(nil)
var _ DatabaseProvider = (*DatabaseContainer)(nil)

// In container.go
var _ LoggerProvider = (*Container)(nil)
var _ APIProvider = (*Container)(nil)
var _ ServiceProvider = (*Container)(nil)
var _ FullProvider = (*Container)(nil)

// In gitsvc.go
func ensureGitIntegrationServiceImplementation() {
    var _ PriorityScheduledGitIntegrationService = (*bitbucketcloud.BitbucketCloudSvc)(nil)
}
```

## Integration Service Architecture

### Base Integration Service
```go
type IntegrationService interface {
    di.ServiceProvider
    ValidateEnvVariables() error
    RunJob() error
}
```

### Specialized Integration Services
```go
// Git Integration Services
type GitIntegrationService interface {
    IntegrationService
    RepoPull() error
    GitActivityPull() error
}

type PriorityScheduledGitIntegrationService interface {
    GitIntegrationService
    GitCodeBreakdownPull() error
}

// Future-ready interfaces
type IssueIntegrationService interface {
    IntegrationService
    IssuePull() error
    IssueActivityPull() error
    IssueMetricsPull() error
}

type CICDIntegrationService interface {
    IntegrationService
    BuildPull() error
    PipelinePull() error
    DeploymentPull() error
}
```

## Usage Examples

### 1. API Services (Client)
```go
// Client only needs API-related dependencies
func NewClient(apiProvider di.APIProvider) *Client {
    return &Client{
        logger:       apiProvider.GetLogger(),
        stateManager: apiProvider.GetStateManager(),
        credentials:  apiProvider.GetCredentials(),
        httpClient:   apiProvider.GetHTTPClient(),
    }
}
```

### 2. Main Services (BitbucketCloudSvc)
```go
// Service needs service, API, and database providers
func NewBitbucketCloudSvc(serviceProvider di.ServiceProvider, apiProvider di.APIProvider, databaseProvider di.DatabaseProvider) *BitbucketCloudSvc {
    return &BitbucketCloudSvc{
        serviceProvider:  serviceProvider,  // For config, validation, etc.
        databaseProvider: databaseProvider, // For database operations
        apiClient:        NewClient(apiProvider),
    }
}

// Interface implementation methods (required for IntegrationService)
func (bcSvc *BitbucketCloudSvc) GetLogger() *shared.CustomLogger {
    return bcSvc.serviceProvider.GetLogger()
}

func (bcSvc *BitbucketCloudSvc) GetConfig() *config.Config {
    return bcSvc.serviceProvider.GetConfig()
}

// Usage in service methods (explicit dependency access)
func (bcSvc *BitbucketCloudSvc) RepoPull() error {
    // Use service provider for logging and config
    bcSvc.serviceProvider.GetLogger().Info("Pulling repositories...")
    
    // Use API provider for API calls
    repos, err := bcSvc.apiClient.GetRepositoriesByWorkspace(workspace.Slug, func(s string) {})
    
    // Use database provider for database operations
    for _, repo := range repos {
        if repoSyncAudit, err := bcSvc.databaseProvider.GetDBQueries().GetRepoSyncAuditByID(context.Background(), repo.Slug); err != nil {
            bcSvc.serviceProvider.GetLogger().Error("Error getting repo sync audit", "error", err)
            return err
        }
        // ... database operations
    }
    return nil
}
```

### 3. Different Main Files
```go
// cmd/datapuller/main.go - needs everything
container := di.NewContainer().
    SetLogger(logger).
    SetConfig(config).
    SetStateManager(stateManager).
    SetCredentials(credentials).
    SetDB(db).
    SetHTTPClient(httpClient)

// Validate ALL dependencies (datapuller needs everything)
if err := container.Validate(); err != nil {
    log.Fatalf("Container validation failed: %v", err)
}

service := NewDataPullerService(container)

// cmd/authsync/main.go - needs only basic dependencies
serviceContainer := di.NewServiceContainer().
    SetLogger(logger).
    SetConfig(config).
    SetStateManager(stateManager).
    SetCredentials(credentials)
    // Note: No DB or HTTP client needed

// Validate only service dependencies (no DB, no HTTP client)
if err := serviceContainer.Validate(); err != nil {
    log.Fatalf("Service validation failed: %v", err)
}

service := NewAuthSyncService(serviceContainer)

// api/main.go - needs only API dependencies
apiContainer := di.NewAPIContainer().
    SetLogger(logger).
    SetStateManager(stateManager).
    SetCredentials(credentials).
    SetHTTPClient(httpClient)
    // Note: No DB or config needed

// Validate only API dependencies (no DB, no config)
if err := apiContainer.Validate(); err != nil {
    log.Fatalf("API validation failed: %v", err)
}

service := NewAPIService(apiContainer)
```

### 4. Integration Service Factory
```go
func GetActiveIntegrationService(container *di.Container) (IntegrationService, error) {
    switch container.Config.ActiveService {
    case config.BitbucketCloudKey:
        // Create specialized containers for different needs
        serviceContainer := container.ToServiceContainer()
        apiContainer := container.ToAPIContainer()
        databaseContainer := container.ToDatabaseContainer()

        return bitbucketcloud.NewBitbucketCloudSvc(serviceContainer, apiContainer, databaseContainer), nil
    default:
        return nil, fmt.Errorf("unsupported service type: %s", container.Config.ActiveService)
    }
}
```

## Benefits

1. **Principle of Least Privilege**: Services only get the dependencies they need
2. **Flexible for Different Mains**: Each main file can use different containers
3. **Maintains Testability**: Easy to create mock providers for testing
4. **Clear Dependencies**: Interface segregation makes dependencies explicit
5. **No Overhead**: No unnecessary dependencies passed around
6. **Easy to Extend**: Add new providers for new dependency combinations
7. **Database Integration**: Services can easily access database operations when needed
8. **Compile-time Safety**: Interface implementation checks ensure compliance
9. **Future-ready**: Extensible architecture for new integration types
10. **Explicit Dependency Access**: Clear where dependencies come from in service methods

## Adding New Dependencies

### Step 1: Define Provider Interface
```go
type CacheProvider interface {
    GetCache() *CacheService
}
```

### Step 2: Add to Specialized Container
```go
type ServiceContainer struct {
    // ... existing fields ...
    Cache *CacheService
}

func (sc *ServiceContainer) GetCache() *CacheService {
    return sc.Cache
}

func (sc *ServiceContainer) SetCache(cache *CacheService) *ServiceContainer {
    sc.Cache = cache
    return sc
}
```

### Step 3: Use in Services
```go
func NewMyService(serviceProvider di.ServiceProvider) *MyService {
    return &MyService{
        cache: serviceProvider.GetCache(),
    }
}
```

## Migration Guide

1. **Update service constructors** to accept specific providers instead of full container
2. **Create specialized containers** in main functions based on needs
3. **Update tests** to use mock providers
4. **Remove unused dependencies** from service constructors
5. **Add interface implementation checks** for compile-time safety

## Best Practices

1. **Use the most specific provider** for each service
2. **Create specialized containers** for different main files
3. **Keep providers focused** - don't mix unrelated dependencies
4. **Use interface composition** to create new provider combinations
5. **Document provider requirements** for each service
6. **Add interface implementation checks** in init() functions
7. **Use explicit dependency access** in service methods (e.g., `serviceProvider.GetLogger()`)
8. **Plan for future extensibility** with interface hierarchies 