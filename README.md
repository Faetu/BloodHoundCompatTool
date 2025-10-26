# BloodHoundCompatTool

BloodHoundCompatTool is a utility designed to ensure backwards compatibility for newer versions of Sharphound Zips, allowing legacy BloodHound to process them correctly.

This tool will read the Sharphound output Zip, and create a modified "compatible" version of it. This new Zip is useable by legacy Bloodhound and also by BloodHound CE.

# Installation

To install BloodHoundCompatTool, you may download precompiled binaries form the realease page.

# Usage

```bash
BloodHoundCompatTool -i <Sharphound ZIP> -o <Output>
```

The `-o` flag for output, is optional. If not set, the fixed zip will be named like the original sharphound zip, with the additional `_Compatible`.

# CAUTION!

While this tool was created with the intent to enable backwards compatibility with legacy versions of BloodHound, minimal effort has been put into ensuring 100% accuracy.
Please be aware:

- There may be `false positives` in the results, expecially when dealing with complex SharpHound exports or specific configurations.
- Edge cases or unsupported data formats might cause unexpected behavior or incomplete data processing.
- This tool should not be relied upon for production environments or critical assessments without further validation.
  Use at your own risk and always verify the results through additional means.

# License

Shield: [![CC BY-NC-SA 4.0][cc-by-nc-sa-shield]][cc-by-nc-sa]

This work is licensed under a
[Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License][cc-by-nc-sa].

[![CC BY-NC-SA 4.0][cc-by-nc-sa-image]][cc-by-nc-sa]

[cc-by-nc-sa]: http://creativecommons.org/licenses/by-nc-sa/4.0/
[cc-by-nc-sa-image]: https://licensebuttons.net/l/by-nc-sa/4.0/88x31.png
[cc-by-nc-sa-shield]: https://img.shields.io/badge/License-CC%20BY--NC--SA%204.0-lightgrey.svg
