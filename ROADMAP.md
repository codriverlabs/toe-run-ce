# TOE Roadmap

This document outlines the planned features and improvements for the TOE (Temporary Observability Engine) project.

## Security & Safety Enhancements

### 1. Command Line Argument Sanitization
- **Goal**: Implement extra sanitization for command line arguments in PowerTool
- **Approach**: Define sub-CRD that will limit the vector of attack
- **Priority**: High
- **Status**: Planned

### 2. PowerToolConfig Argument Enforcement
- **Goal**: Enforce acceptable command line arguments in PowerToolConfig
- **Approach**: Implement validation rules and whitelisting mechanisms
- **Priority**: High
- **Status**: Planned

## Reliability & Resource Management

### 3. Signal Handling for PowerTool
- **Goal**: Add handling and trapping termination signals for each PowerTool
- **Approach**: Implement graceful shutdown to save work when collection pod gets interrupted
- **Priority**: Medium
- **Status**: Planned

### 4. Default Resource Specifications
- **Goal**: PowerToolConfig should specify default CPU and memory requests and limits
- **Approach**: Define resource templates and defaults in configuration
- **Priority**: Medium
- **Status**: Planned

## Data Processing & Visualization

### 5. PowerToolProcessing CRD
- **Goal**: Define PowerToolProcessingCRD for each collector
- **Features**:
  - Specify frequency of processing collected data
  - Define how to expose processed data
  - For aperf: HTML reports exposed via ingress
  - Processing using LLMs/SLMs for automated RCA (Root Cause Analysis)
- **Priority**: High
- **Status**: Planned

## Development & Deployment Optimization

### 6. Helm Charts Optimization
- **Goal**: Optimize helm charts generation
- **Approach**: Streamline chart structure and improve templating
- **Priority**: Medium
- **Status**: Planned

### 7. GitHub Actions Artifacts
- **Goal**: Optimize generation of artifacts via GitHub Actions
- **Approach**: Improve CI/CD pipeline efficiency and artifact management
- **Priority**: Medium
- **Status**: Planned

## User Experience & Adoption

### 8. Interactive Tutorial
- **Goal**: Implement "Hello PowerTool" interactive tutorial
- **Approach**: Create step-by-step guided experience for new users
- **Priority**: Low
- **Status**: Planned

### 9. Community Engagement
- **Goal**: Have lots of workshops and coffees
- **Approach**: Organize community events, workshops, and informal meetups
- **Priority**: Ongoing
- **Status**: Planned

## Implementation Timeline

- **Phase 1**: Security & Safety (Items 1-2)
- **Phase 2**: Reliability & Resources (Items 3-4)
- **Phase 3**: Data Processing (Item 5)
- **Phase 4**: Optimization (Items 6-7)
- **Phase 5**: User Experience (Items 8-9)

## Contributing

Contributions to any of these roadmap items are welcome. Please see our contributing guidelines and feel free to open issues or pull requests for discussion.
