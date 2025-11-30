# Thread Safety Solution for Acquire Pattern

## The Real Problem

Your current Acquire pattern has a race condition:

```go
var bitbucketCloudSvc *BitbucketCloudSvc

func AcquireBitbucketCloudSvc() *BitbucketCloudSvc {
    if bitbucketCloudSvc == nil {  // Race condition here!
        bitbucketCloudSvc = NewBitbucketCloudSvc(...)
    }
    return bitbucketCloudSvc
}
```

## Simple Solution

Replace the global variable with a thread-safe singleton:

```go
// Instead of: var bitbucketCloudSvc *BitbucketCloudSvc
var bitbucketCloudSvc = di.NewThreadSafeSingleton(func() *BitbucketCloudSvc {
    // This creator function only runs once, thread-safely
    customLogger := shared.AcquireCustomLogger()
    cfg := config.AcquireConfig()
    statemanager := statemanager.AcquireStateManager()
    credentials := credservice.AcquireCredentials()
    dbQuerier := dbsetup.AcquireQueries()
    client := AcquireClient()
    
    return NewBitbucketCloudSvc(customLogger, statemanager, credentials, cfg, dbQuerier, client)
})

func AcquireBitbucketCloudSvc() *BitbucketCloudSvc {
    return bitbucketCloudSvc.Acquire()
}
```

## Benefits

✅ **Thread Safe** - No race conditions  
✅ **Simple** - Just wraps your existing pattern  
✅ **Testable** - `Reset()` and `SetForTesting()` methods  
✅ **Fast** - Fast path for existing instances  
✅ **Minimal Changes** - Keep your existing logic  

## Usage

### Production
```go
service := bitbucketcloud.AcquireBitbucketCloudSvc()
```

### Testing
```go
func TestService() {
    // Reset for clean state
    bitbucketCloudSvc.Reset()
    
    // Set mock for testing
    mockService := NewBitbucketCloudSvc(mockLogger, mockStateManager, ...)
    bitbucketCloudSvc.SetForTesting(mockService)
    
    // Test with mock
    service := AcquireBitbucketCloudSvc()
    // ... test logic
    
    // Clean up
    bitbucketCloudSvc.Reset()
}
```

## Migration

1. Replace `var service *Service` with `var service = di.NewThreadSafeSingleton(...)`
2. Update `Acquire*()` methods to call `service.Acquire()`
3. Update tests to use `Reset()` and `SetForTesting()`

That's it! No complex dependency managers, no provider patterns, just thread safety. 