# Specification Format Guidelines

All specification sections and proposals in this project MUST follow the IETF RFC/Internet-Draft format structure. This ensures technical rigor, clarity, and completeness.

### 1. **Introduction** (Required)
- **Purpose**: Provide context for WHY this specification exists
- **Content**:
  - Current state of the system
  - Problem being solved or capability being added
  - Why existing approaches are insufficient
  - Brief overview of the solution approach
- **Length**: 2-4 paragraphs

### 2. **Requirements Notation** (If using normative language)
- If using RFC 2119 keywords (MUST, SHOULD, MAY, etc.), include this section:
  > "The key words 'MUST', 'MUST NOT', 'REQUIRED', 'SHALL', 'SHALL NOT', 'SHOULD', 'SHOULD NOT', 'RECOMMENDED', 'MAY', and 'OPTIONAL' in this document are to be interpreted as described in RFC 2119."

### 3. **Terminology** (Required if domain-specific terms used)
- Define all domain-specific terms used in the specification
- Clarify any ambiguous terms (e.g., "client" vs "application")
- Use format: `Term`: Definition
- **Example**:
  ```
  "Flow": A directed graph of executable nodes that transforms data
  "Execution": An instance of a flow being processed by the system
  "Connection": Authenticated credentials for external service integration
  ```

### 4. **Concepts** (Required for complex features)
- Explain core concepts that underpin the specification
- Break down into subsections (6.1, 6.2, etc.) for distinct concepts
- Use clear, accessible language
- Include examples where helpful

### 5. **Core Technical Sections** (Required - numbered appropriately)
Structure these based on your domain:
- **API Specifications**: Request/response formats, endpoints, parameters
- **Data Models**: Schema definitions, relationships, constraints
- **Behaviors**: Expected system behaviors, state transitions
- **Algorithms**: Step-by-step processing logic

**For each technical element, specify**:
- Required vs optional components (use MUST/MAY/SHOULD)
- Data types and formats
- Valid ranges or values
- Default behaviors
- Error conditions

### 6. **Examples** (Highly Recommended)
- Provide concrete, runnable examples
- Show both typical cases AND edge cases
- Use realistic data
- Include complete request/response cycles where applicable
- Format code blocks clearly with syntax highlighting hints

### 7. **Security Considerations** (Required)
- Identify ALL security implications of the specification
- Address:
  - Authentication/authorization impacts
  - Data exposure risks
  - Injection vulnerabilities
  - Rate limiting/DoS concerns
  - Cryptographic requirements
- Use normative language (MUST/SHOULD) for security requirements

### 8. **Privacy Considerations** (Required if handling user data)
- Address data collection, storage, and retention
- Identify PII or sensitive data handling
- Specify data access controls
- Consider regulatory compliance (GDPR, etc.)

### 9. **Error Handling** (Required for API/protocol specs)
- Define all error conditions
- Specify error codes/types
- Describe expected error responses
- Document error recovery procedures

### 10. **Migration/Upgrade Path** (Required if changing existing functionality)
- Backward compatibility strategy
- Migration steps for existing data/code
- Deprecation timeline if applicable
- Rollback procedures

### 11. **References** (Required if citing external sources)
Split into:
- **Normative References**: Specifications/standards that MUST be consulted
- **Informative References**: Helpful background reading

Format: `[ShortName]` Author, "Title", Date, URL

## Writing Style Requirements

1. **Precision**: Use exact, unambiguous language
2. **Normative Language**: When specifying requirements, use RFC 2119 keywords consistently
3. **Active Voice**: Prefer "The system MUST validate..." over "Validation must be performed..."
4. **Present Tense**: Describe behavior as "returns" not "will return"
5. **Structured Lists**: Use bullet points or numbered lists for multiple related items
6. **Examples**: Mark clearly as "Example:" or use indented code blocks
7. **Consistency**: Use the same term for the same concept throughout
8. **No Emojis**: Do NOT use emojis in specifications or documentation - use clear, professional language instead

## Section Numbering

- Use decimal numbering: 1, 1.1, 1.1.1, etc.
- Number ALL major sections
- Subsections inherit parent numbering
- References and Acknowledgments may be unnumbered

## What NOT to Include

- **Implementation details**: Focus on WHAT and WHY, not HOW (save for design.md)
- **Vague requirements**: Every MUST/SHOULD must be testable
- **Assumptions**: State all prerequisites explicitly
- **Marketing language**: Keep technical and objective
