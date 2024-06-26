# AutoDev

AutoDev is an AI-powered development tool that integrates with GitHub repositories and uses the Greptile API for advanced code search and analysis.

## Features

- Repository indexing
- Code querying
- Semantic search
- GitHub integration

## Prerequisites

- Go 1.16+
- Greptile API key
- GitHub access token

## Setup

1. Clone the repository:

   ```
   git clone https://github.com/OpenAgentsInc/autodev.git
   cd autodev
   ```

2. Install dependencies:

   ```
   go mod tidy
   ```

3. Create a `.env` file in the project root:

   ```
   GREPTILE_API_KEY=your_greptile_api_key
   GITHUB_TOKEN=your_github_token
   ```

4. Build the project:
   ```
   go build
   ```

## Usage

1. Start the server:

   ```
   ./autodev
   ```

2. Open a web browser and navigate to `http://localhost:8080`

3. Use the web interface to interact with your GitHub repositories through AutoDev's features.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the [AGPL-3.0-or-later](LICENSE).
