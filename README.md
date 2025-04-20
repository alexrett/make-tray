# MakeTray

`MakeTray` is a lightweight utility designed to simplify the creation of system tray applications. It provides a streamlined interface for managing system tray icons and menus across platforms.

## Features

- Run make commands directly from you Mac OS bar

## Installation

Clone the repository:

```bash
git clone https://github.com/alexrett/make-tray.git
cd make-tray
```

Install dependencies:

```bash
go mod tidy
```

Build and install:

```bash
make build
```

Application will be installed into `/Application/MakeTray.app` 

## Usage

- Create in iCloud root folder `Makefile` (use `make createMake` command for initialize)
- Run application MakeTray.app
- Edit your Makefile stored in iCloud with your needs
- Enjoy


## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).