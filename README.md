# SIP Parser

## Overview
The SIP Parser is a Go-based application designed to read and validate text files containing SIP (Session Initiation Protocol) messages. This tool ensures that the SIP messages conform to the expected format and standards.

## Features
- Reads SIP messages from text files.
- Validates the structure and content of SIP messages.
- Provides feedback on the validity of the messages.

## Installation
1. Ensure you have Go installed on your machine. You can download it from [here](https://golang.org/dl/).
2. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/sip-parser.git
    ```
3. Navigate to the project directory:
    ```sh
    cd sip-parser
    ```
4. Build the project:
    ```sh
    go build -o sip-parser
    ```

## Usage
1. Run the SIP Parser with a text file containing SIP messages:
    ```sh
    ./sip-parser path/to/sip-messages.txt
    ```
2. The parser will output whether the messages are valid or not.

## Example
```sh
./sip-parser examples/sip-messages.txt
```

Output:
```
Message 1: Valid
Message 2: Invalid - Missing Via header
...
```

## Contributing

1. Fork the repository.
2. Create a new branch (git checkout -b feature-branch).
3. Make your changes.
4. Commit your changes (git commit -am 'Add new feature').
5. Push to the branch (git push origin feature-branch).
6. Create a new Pull Request.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contact

For any questions or suggestions, please open an issue.