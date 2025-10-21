# Profiler Images

This directory contains Dockerfiles for various profiling tools used by the toe controller.

## Available Profilers

### aperf
- **Base Image**: Amazon Linux 2023 Minimal
- **Tool**: AWS aperf v0.1.18-alpha (AWS performance profiler)
- **Usage**: Profiles applications using Linux perf via PID attachment
- **Architecture**: Supports x86_64 and aarch64

### aprof
- **Base Image**: Amazon Linux 2023 Minimal
- **Tool**: DataDog aprof (Go profiler)
- **Usage**: Profiles Go applications via PID attachment

## Building Images

```bash
# Build aperf profiler image
docker build -t toe-profiler-aperf:latest profilers/aperf/

# Build aprof profiler image
docker build -t toe-profiler-aprof:latest profilers/aprof/

# Push to registry
docker tag toe-profiler-aperf:latest <registry>/toe-profiler-aperf:latest
docker push <registry>/toe-profiler-aperf:latest
```

## Environment Variables

### aperf
- `DURATION`: Profiling duration in seconds (default: 30)
- `PROFILE_TYPE`: Profile type - cpu, memory, etc. (default: cpu)
- `TARGET_PID`: Process ID to profile (default: 1)
- `OUTPUT_PATH`: Output file path (default: /tmp/profile.json)

### aprof
- `DURATION`: Profiling duration (default: 30s)
- `FORMAT`: Output format (default: pprof)
- `SAMPLE_RATE`: Sampling rate (default: 100)
- `TARGET_PID`: Process ID to profile (default: 1)
- `OUTPUT_PATH`: Output file path (default: /tmp/profile.pprof)
