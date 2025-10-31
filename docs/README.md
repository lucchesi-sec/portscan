# PortScan Documentation

Welcome to the PortScan documentation. This directory contains comprehensive documentation covering architecture, design decisions, security considerations, and development guidelines.

## 📚 Documentation Structure

The repository currently ships these primary documents:

- `ARCHITECTURE.md` – end-to-end system architecture and design choices
- `DEVELOPER_GUIDE.md` – local workflows, tooling, and contribution tips
- `DOCUMENTATION_SUMMARY.md` – status overview and quick index
- `MAINTENANCE.md` – release and dependency maintenance procedures
- `SECURITY_SCANNING.md` – how to run static/dynamic security tooling
- `README.md` – entry point for the documentation folder

Generated artefacts from `scripts/generate-docs.go` are written to `docs/generated/` after running `make generate-docs`.

## 🏗️ Architecture Documentation

The architecture documentation provides a comprehensive view of the PortScan system:

### [Main Architecture Overview](./architecture/README.md)
Complete system architecture including:
- System context and boundaries
- Container and component architecture
- Data flow and sequence diagrams
- Quality attributes and performance characteristics

### [C4 Model Diagrams](./architecture/c4-model/)
Hierarchical architecture views:
- **Context**: System in its environment
- **Container**: High-level technology choices
- **Component**: Internal structure
- **Code**: Implementation details

### [Architecture Decision Records](./architecture/adrs/)
Key architectural decisions:
- [ADR-001](./architecture/adrs/ADR-001-worker-pool-pattern.md): Worker Pool Pattern
- [ADR-002](./architecture/adrs/ADR-002-channel-based-events.md): Channel-Based Events
- [ADR-003](./architecture/adrs/ADR-003-bubble-tea-tui.md): Bubble Tea TUI
- [ADR-004](./architecture/adrs/ADR-004-ndjson-default.md): NDJSON Default Format
- [ADR-005](./architecture/adrs/ADR-005-rate-limiting.md): Rate Limiting

### [Security Architecture](./architecture/security/)
Security considerations and threat model:
- [Security Overview](./architecture/security/README.md)
- [Threat Model](./architecture/security/threat-model.md)
- Trust boundaries and controls
- Compliance considerations

## 🔧 Building Documentation

### Prerequisites (optional helpers)
```bash
# Diagram rendering (optional)
brew install plantuml graphviz

# Mermaid + Markdown linting (optional)
npm install -g @mermaid-js/mermaid-cli markdownlint-cli

# Additional generators (optional)
go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
```

### Generate Documentation
```bash
# Generate code documentation snapshot
make generate-docs
```

The optional tooling above can be installed when you need diagram exports or Markdown linting, but it is not required for day-to-day development.

### Automated Generation
Documentation is automatically generated on:
- Push to main branch
- Pull requests
- Manual workflow dispatch

See [.github/workflows/docs.yml](../.github/workflows/docs.yml) for CI/CD configuration.

## 📖 Documentation Types

### 1. Architecture Documentation
- **Purpose**: System design and structure
- **Audience**: Developers, architects
- **Format**: Markdown with embedded diagrams
- **Location**: `docs/architecture/`

### 2. API Documentation
- **Purpose**: Code-level documentation
- **Audience**: Developers
- **Format**: Auto-generated from code
- **Location**: `docs/api/`

### 3. User Guides
- **Purpose**: Usage instructions
- **Audience**: End users
- **Format**: Markdown tutorials
- **Location**: `docs/guides/`

### 4. Generated Documentation
- **Purpose**: Auto-extracted from code
- **Audience**: Developers
- **Format**: Multiple (MD, JSON, HTML)
- **Location**: `docs/generated/`

## 🎨 Diagram Types

### PlantUML Diagrams
- C4 model diagrams
- Sequence diagrams
- State machines
- Component diagrams

### Mermaid Diagrams
- Embedded in markdown
- Flowcharts
- Entity relationships
- Gantt charts

## 📝 Documentation Standards

### Markdown Guidelines
- Use ATX headers (`#`)
- Include TOC for long documents
- Use code blocks with language hints
- Add diagrams where helpful

### Diagram Guidelines
- Keep diagrams focused
- Use consistent styling
- Include legends where needed
- Version control source files

### ADR Guidelines
- One decision per ADR
- Include context and consequences
- Never modify accepted ADRs
- Link related ADRs

## 🚀 Quick Links

- [Architecture Overview](./architecture/README.md)
- [Security Documentation](./architecture/security/README.md)
- [ADR Index](./architecture/adrs/README.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Main README](../README.md)

## 📊 Documentation Coverage

Current documentation includes:
- ✅ Complete C4 architecture model
- ✅ 5 Architecture Decision Records
- ✅ Security architecture and threat model
- ✅ Data flow and sequence diagrams
- ✅ Automated generation pipeline
- ✅ Multiple output formats

## 🔄 Keeping Documentation Updated

1. **Code Changes**: Update relevant documentation
2. **Architecture Changes**: Create/update ADRs
3. **New Features**: Add to architecture docs
4. **Security Changes**: Update threat model
5. **API Changes**: Regenerate API docs

## 📬 Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on contributing to documentation.

### Documentation Contributions Welcome
- Architecture improvements
- Additional diagrams
- User guides
- Examples and tutorials
- Corrections and clarifications

## 📄 License

Documentation is provided under the same license as the project. See [LICENSE](../LICENSE) for details.
