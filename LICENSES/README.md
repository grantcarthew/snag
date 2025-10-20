# Third-Party Licenses

`snag` is licensed under the Mozilla Public License 2.0 (see ../LICENSE).

This directory contains the licenses for third-party dependencies used in snag.

## Dependencies

### go-rod/rod (MIT License)

- **Purpose**: Chrome DevTools Protocol library for browser automation
- **Repository**: https://github.com/go-rod/rod
- **License**: MIT License
- **License File**: rod.LICENSE

### urfave/cli (MIT License)

- **Purpose**: CLI framework for building command-line applications
- **Repository**: https://github.com/urfave/cli
- **License**: MIT License
- **License File**: urfave-cli.LICENSE

### JohannesKaufmann/html-to-markdown (MIT License)

- **Purpose**: HTML to Markdown conversion library
- **Repository**: https://github.com/JohannesKaufmann/html-to-markdown
- **License**: MIT License
- **License File**: html-to-markdown.LICENSE

## License Compatibility

All dependencies use the MIT License, which is compatible with MPL 2.0:

- MIT is a permissive license allowing commercial use, modification, and distribution
- MIT-licensed code can be included in MPL 2.0 projects
- Proper attribution is maintained in this directory

## Generating License Files

To verify or update these licenses:

```bash
# Download license files from repositories
curl -L https://raw.githubusercontent.com/go-rod/rod/main/LICENSE -o rod.LICENSE
curl -L https://raw.githubusercontent.com/urfave/cli/main/LICENSE -o urfave-cli.LICENSE
curl -L https://raw.githubusercontent.com/JohannesKaufmann/html-to-markdown/main/LICENSE -o html-to-markdown.LICENSE
```

## Acknowledgments

We thank the maintainers and contributors of these excellent open-source projects.
