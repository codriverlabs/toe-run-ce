# Unit Test Improvement Plan

## Current Status

### Coverage Summary
- **Storage Manager**: 72.2% ✅
- **Label Matching**: 100% ✅
- **Env Vars Builder**: 60.0% ⚠️
- **Collector Server**: 0.0% ❌
- **Auth Module**: 0.0% ❌
- **Overall**: 15.7%

### Completed Tests
- ✅ `pkg/collector/storage/manager_test.go`
- ✅ `internal/controller/label_matching_test.go`
- ✅ `internal/controller/env_vars_test.go`

---

## Priority 1: Collector Server Tests (High Priority)

### Objective
Achieve 70%+ coverage for `pkg/collector/server/server.go`

### Test File
`pkg/collector/storage/server_test.go`

### Test Cases

#### 1. NewServer Tests
```go
TestNewServer
├── valid configuration
├── invalid storage path
├── empty date format
├── nil k8s client
└── TLS configuration
```

**Coverage Target**: 80%

#### 2. handleProfile Tests
```go
TestHandleProfile
├── successful profile upload
├── missing Authorization header
├── invalid token
├── missing X-PowerTool-Job-ID header
├── missing X-PowerTool-Namespace header
├── missing X-PowerTool-Filename (uses default)
├── empty X-PowerTool-Matching-Labels (defaults to unknown)
├── large file upload
├── concurrent uploads
└── storage failure handling
```

**Coverage Target**: 85%

#### 3. HTTP Method Tests
```go
TestHandleProfile_HTTPMethods
├── POST request (success)
├── GET request (405 Method Not Allowed)
├── PUT request (405 Method Not Allowed)
└── DELETE request (405 Method Not Allowed)
```

**Coverage Target**: 100%

#### 4. Metadata Extraction Tests
```go
TestHandleProfile_MetadataExtraction
├── all headers present
├── optional headers missing
├── header value sanitization
└── special characters in headers
```

**Coverage Target**: 90%

### Implementation Approach

**Mock Dependencies:**
- Mock K8s client for token validation
- Mock storage manager for file operations
- Use `httptest.NewRecorder()` for HTTP testing

**Example Test Structure:**
```go
func TestHandleProfile_Success(t *testing.T) {
    // Setup
    mockStorage := &mockStorageManager{}
    mockAuth := &mockTokenValidator{}
    server := &Server{
        storage: mockStorage,
        auth: mockAuth,
    }
    
    // Create request
    body := bytes.NewBufferString("test data")
    req := httptest.NewRequest("POST", "/api/v1/profile", body)
    req.Header.Set("Authorization", "Bearer valid-token")
    req.Header.Set("X-PowerTool-Job-ID", "test-job")
    req.Header.Set("X-PowerTool-Namespace", "default")
    req.Header.Set("X-PowerTool-Matching-Labels", "app-nginx")
    req.Header.Set("X-PowerTool-Filename", "output.txt")
    
    // Execute
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(server.handleProfile)
    handler.ServeHTTP(rr, req)
    
    // Assert
    if rr.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", rr.Code)
    }
}
```

### Estimated Effort
- **Time**: 4-6 hours
- **Complexity**: Medium
- **Dependencies**: Mock interfaces for storage and auth

---

## Priority 2: Auth Module Tests (High Priority)

### Objective
Achieve 70%+ coverage for `pkg/collector/auth/`

### Test File
`pkg/collector/auth/validator_test.go`

### Test Cases

#### 1. Token Validation Tests
```go
TestValidateToken
├── valid token
├── expired token
├── invalid signature
├── malformed token
├── empty token
├── token without required claims
└── token from wrong service account
```

**Coverage Target**: 85%

#### 2. K8s Client Integration Tests
```go
TestK8sTokenValidator
├── successful token review
├── K8s API error handling
├── network timeout
├── unauthorized token
└── service account not found
```

**Coverage Target**: 75%

#### 3. Token Manager Tests
```go
TestGenerateToken
├── successful token generation
├── custom duration
├── minimum duration enforcement (10 minutes)
├── service account not found
└── K8s API error
```

**Coverage Target**: 80%

### Implementation Approach

**Mock K8s Client:**
```go
type mockK8sClient struct {
    kubernetes.Interface
    tokenReview func(context.Context, *authv1.TokenReview) (*authv1.TokenReview, error)
}
```

**Example Test:**
```go
func TestValidateToken_Success(t *testing.T) {
    mockClient := &mockK8sClient{
        tokenReview: func(ctx context.Context, tr *authv1.TokenReview) (*authv1.TokenReview, error) {
            tr.Status.Authenticated = true
            tr.Status.User.Username = "system:serviceaccount:toe-system:toe-sdk-collector"
            return tr, nil
        },
    }
    
    validator := NewK8sTokenValidator(mockClient, "toe-sdk-collector")
    userInfo, err := validator.ValidateToken(context.Background(), "valid-token")
    
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    if userInfo.Username != "system:serviceaccount:toe-system:toe-sdk-collector" {
        t.Errorf("unexpected username: %v", userInfo.Username)
    }
}
```

### Estimated Effort
- **Time**: 3-5 hours
- **Complexity**: Medium-High
- **Dependencies**: K8s client mocking

---

## Priority 3: Integration Tests (Medium Priority)

### Objective
End-to-end testing of complete workflows

### Test File
`test/integration/collector_integration_test.go`

### Test Scenarios

#### 1. Complete Profile Upload Flow
```go
TestCompleteProfileUpload
├── Controller creates ephemeral container
├── Power-tool sends profile to collector
├── Collector validates token
├── Collector saves to hierarchical path
└── Verify file exists at correct location
```

**Coverage Target**: Full workflow

#### 2. Multi-Component Interaction
```go
TestMultiComponentInteraction
├── Multiple PowerTools targeting same pod
├── Concurrent profile uploads
├── Different label selectors
└── Various date formats
```

**Coverage Target**: Full workflow

#### 3. Error Scenarios
```go
TestErrorScenarios
├── Storage full
├── Invalid token
├── Network failures
├── Malformed requests
└── Recovery mechanisms
```

**Coverage Target**: Full workflow

### Implementation Approach

**Use envtest for K8s:**
```go
func TestMain(m *testing.M) {
    testEnv = &envtest.Environment{
        CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
    }
    
    cfg, err := testEnv.Start()
    // ... setup
    
    code := m.Run()
    testEnv.Stop()
    os.Exit(code)
}
```

**Example Integration Test:**
```go
func TestCompleteProfileUpload(t *testing.T) {
    // Setup collector
    collector := startCollector(t)
    defer collector.Stop()
    
    // Create PowerTool
    powerTool := createPowerTool(t, "test-profile")
    
    // Simulate power-tool upload
    token := generateToken(t)
    uploadProfile(t, collector.URL, token, "test-data")
    
    // Verify file exists
    expectedPath := "/data/default/app-test/test-profile/2025/10/30/output.txt"
    verifyFileExists(t, expectedPath)
}
```

### Estimated Effort
- **Time**: 6-8 hours
- **Complexity**: High
- **Dependencies**: envtest, test fixtures

---

## Additional Improvements

### 4. Controller Tests Enhancement

#### Improve Existing Coverage
```go
TestBuildPowerToolEnvVars_ToolArgs
├── JSON args parsing
├── invalid JSON handling
├── nested args
└── special characters in args
```

**Target**: 60% → 80%

### 5. End-to-End Tests

#### Real Cluster Testing
```go
TestE2E_RealCluster
├── Deploy collector
├── Deploy PowerToolConfig
├── Create target pod
├── Create PowerTool
├── Wait for completion
└── Verify results
```

**Target**: Full workflow validation

---

## Testing Best Practices

### 1. Test Organization
```
pkg/collector/
├── server/
│   ├── server.go
│   └── server_test.go
├── auth/
│   ├── validator.go
│   └── validator_test.go
└── storage/
    ├── manager.go
    └── manager_test.go
```

### 2. Mock Interfaces
```go
// Define interfaces for mocking
type StorageManager interface {
    SaveProfile(io.Reader, ProfileMetadata) error
}

type TokenValidator interface {
    ValidateToken(context.Context, string) (*UserInfo, error)
}
```

### 3. Test Helpers
```go
// test/helpers/helpers.go
func CreateTestPowerTool(name string) *toev1alpha1.PowerTool
func CreateTestPod(name string, labels map[string]string) *corev1.Pod
func GenerateTestToken() string
```

### 4. Table-Driven Tests
```go
tests := []struct {
    name    string
    input   X
    want    Y
    wantErr bool
}{
    // test cases
}
```

---

## Coverage Goals

### Short Term (1-2 weeks)
- ✅ Storage Manager: 72.2%
- ✅ Label Matching: 100%
- 🎯 Collector Server: 70%+
- 🎯 Auth Module: 70%+
- 🎯 Overall: 25%+

### Medium Term (1 month)
- 🎯 Controller: 70%+
- 🎯 Integration Tests: Basic coverage
- 🎯 Overall: 40%+

### Long Term (3 months)
- 🎯 All critical paths: 80%+
- 🎯 E2E tests: Complete workflows
- 🎯 Overall: 60%+

---

## Implementation Timeline

### Week 1-2: High Priority Tests
- [ ] Collector Server Tests
- [ ] Auth Module Tests
- [ ] Mock interfaces setup

### Week 3-4: Integration Tests
- [ ] Basic integration tests
- [ ] Multi-component tests
- [ ] Error scenario tests

### Week 5-6: Enhancement & Polish
- [ ] Improve controller coverage
- [ ] Add E2E tests
- [ ] Documentation updates

---

## Success Metrics

### Quantitative
- Coverage: 15.7% → 60%+
- Test count: 30 → 150+
- Test execution time: <30s

### Qualitative
- All critical paths tested
- Clear test documentation
- Easy to add new tests
- Fast feedback loop

---

## Resources Needed

### Tools
- `go test` - Built-in testing
- `envtest` - K8s testing framework
- `httptest` - HTTP testing
- `gomock` - Mock generation (optional)

### Documentation
- Go testing best practices
- K8s client-go testing
- HTTP handler testing patterns

### Time Investment
- Initial setup: 2-3 days
- Test implementation: 2-3 weeks
- Maintenance: Ongoing

---

## Notes

- Focus on critical paths first
- Use mocks to isolate components
- Keep tests fast and reliable
- Document test scenarios
- Review coverage regularly
- Refactor as needed

---

## References

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Kubernetes Testing Guide](https://kubernetes.io/docs/reference/using-api/client-libraries/)
- [HTTP Testing in Go](https://golang.org/pkg/net/http/httptest/)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
