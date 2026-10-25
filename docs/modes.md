# Calculator modes

The app has four modes sharing one window, one display component and one history/settings system. Each mode has its own accent: Normal blue, Scientific violet, Developer emerald, Finance amber. The accent drives tabs, focus rings, operator keys, the result glow and the display's top hairline.

Common behaviour:

- The display shows the live expression on the top line and the result below.
- `Enter` confirms a calculation; `Escape` clears; `Backspace` deletes.
- Confirmed results are written to history with the mode name.
- Errors appear in the display (or below the form in Finance) and never enter history.
- All animations are disabled under `prefers-reduced-motion: reduce`.

## Normal

Normal mode is the everyday arithmetic surface. It evaluates in TypeScript on every keystroke, so the preview above the result updates instantly with no backend round-trip.

### Input model

Buttons: `AC`, `⌫`, `%`, `÷`, digits, `×`, `−`, `+`, `±`, `.`, `=` and a memory bar (`MC`, `MR`, `M+`, `M−`).

The display expression uses typographic glyphs (`×`, `÷`, `−`) while the evaluator works on ASCII operators. Parentheses can be typed from a hardware keyboard even though the keypad does not include them.

Numbers are entered one decimal point at a time; a second decimal point is ignored. A leading `0` is replaced by the next digit. After `=` the next digit starts a new expression, while an operator continues from the result.

### Grammar and precedence

```
expression := term (('+' | '-') term)*
term       := unary (('*' | '/') unary)*
unary      := ('+' | '-') unary | primary
primary    := number | '(' expression ')'
postfix    := primary '%'
```

- `*` and `/` bind tighter than `+` and `-`.
- Unary minus applies to the operand that follows.
- `%` is postfix.
- Division by zero reports `Can't divide by zero`.
- Non-finite results report `Result is out of range`.
- `-0` is normalized to `0`.

### Percentage semantics

- `200 + 10%` → `220` — the percent is taken relative to the left operand.
- `200 − 10%` → `180`.
- `50%` → `0.5` when used alone.
- `200 × 10%` → `20`.
- `%` after an operator or an open parenthesis is ignored.

### Formatting

Results use up to 12 significant digits with grouping (`1,234,567.89`). Very large or very small magnitudes switch to exponential notation (`1.234e+12`, `1.23e-10`). The displayed font size steps down as the string grows so long results remain visible.

### Keyboard

| Key | Action |
| --- | --- |
| `0`–`9`, `.` or `,` | Enter digits / decimal point |
| `+`, `-`, `*` or `x`, `/` | Operators |
| `%` | Percent |
| `(` `)` | Parentheses |
| `Enter` or `=` | Evaluate |
| `Backspace` | Delete last character |
| `Escape` or `Delete` | Clear |
| `Ctrl`/`Cmd`/`Alt` + key | Ignored, so OS shortcuts keep working |

## Scientific

Scientific mode evaluates in Go through the shared Pratt parser. It supports live expression editing, a debounced preview, angle units and a second-function layer.

### Input model

- Digits, `.`, `±`, `⌫`, `AC`, parentheses, `+ − × ÷`.
- `x²`/`x³` append `^2`/`^3` to the current value; `xʸ` inserts `^`.
- `!` appends factorial.
- `mod` appends `%`; with 2nd active it inserts `abs(`.
- `2nd` swaps trig/log/root/constant keys to their inverse or alternate forms.
- `DEG`/`RAD`/`GRAD` cycles the angle unit and persists it.
- The display shows a small `( n` chip while parentheses are left open.

Function keys insert a name and an opening parenthesis; the parser also accepts function calls without parentheses (`sin 30`). Implicit multiplication is inserted when a value is followed by a function, constant or parenthesis, so `2π` and `3sin(30)` behave as expected.

### Grammar

| Feature | Syntax | Notes |
| --- | --- | --- |
| Arithmetic | `+ - * / % ^` | `^` is right-associative; `%` is modulo here |
| Unary | `-x`, `+x` | `-2^2 = -4`; use `(-2)^2` for the other reading |
| Factorial | `x!` or `fact(x)` | Integer 0–1000, computed with `math/big` |
| Parentheses | `( … )` | Missing closing parenthesis is an error |
| Implicit product | `2(3+4)`, `2pi`, `3sin(0.5)` | Resolved at multiplication precedence |

### Functions

| Group | Functions |
| --- | --- |
| Trigonometric | `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `atan2(y, x)` |
| Hyperbolic | `sinh`, `cosh`, `tanh`, `asinh`, `acosh`, `atanh` |
| Logarithmic / exponential | `ln`, `log` (base 10), `log2`, `exp`, `exp10` |
| Powers and roots | `sqrt`, `cbrt`, `root(x, n)`, `pow(x, y)`, `hypot(x, y)` |
| Rounding | `round`, `floor`, `ceil`, `trunc` |
| Utility | `abs`, `sign`, `mod(x, y)`, `min`, `max`, `fact` |

Constants: `pi`, `tau`, `e`, `phi`.

### Angle units

The selected unit (`deg`, `rad`, `grad`) affects trigonometric functions and their inverses only. It is persisted in settings and restored at startup.

### Second-function layer

| Default | 2nd |
| --- | --- |
| `sin` / `cos` / `tan` | `sin⁻¹` / `cos⁻¹` / `tan⁻¹` |
| `ln` / `log` | `eˣ` / `10ˣ` |
| `x²` / `√` | `x³` / `∛` |
| `π` / `mod` | `τ` / `abs` |

The display shows a `2nd` badge while the layer is active.

### Preview and errors

A 160 ms debounce evaluates the expression and shows a preview result without adding history. `Enter` performs the final evaluation, animates the result, displays `expression =`, and records history. Invalid input leaves the last valid result in place; domain errors such as `sqrt(-1)`, `ln(0)`, `1/0` or `tan` overflow report a readable message. Division by zero and unknown identifiers are reported by the engine.

### Keyboard

| Key | Action |
| --- | --- |
| `0`–`9`, `.` or `,` | Digits / decimal point |
| `+`, `-`, `*` or `x`, `/`, `^`, `%` | Operators (modulo) |
| `!` | Factorial |
| `(` `)` | Parentheses |
| `Enter` or `=` | Evaluate and record history |
| `Backspace` | Delete last character |
| `Escape` or `Delete` | Clear |

## Developer

Developer mode does fixed-width programmer arithmetic. Every value is a word of 8, 16, 32 or 64 bits, interpreted as signed or unsigned.

### Value model

The display shows the active base large, with simultaneous readouts for all four bases:

- `HEX` — zero-padded to the word width (`000000F0` for 8-bit, eight digits for 32-bit).
- `DEC` — unsigned, or two's-complement signed when signed mode is on.
- `OCT` — zero-padded octal.
- `BIN` — zero-padded binary; it wraps across two lines for 32/64-bit widths.

Clicking a readout selects that base for entry. The keypad disables digits that are invalid in the selected base (for example `A`–`F` in decimal). Octal and binary entry restrict digits to 0–7 and 0–1.

The bit grid shows every bit from MSB to LSB; clicking a bit flips it and refreshes all readouts. Values are masked to the selected width on every operation, so overflow wraps.

### Word size and signedness

- Width: 8, 16, 32 or 64 bits.
- Signed: reinterprets the raw bits using two's complement; the DEC readout changes, while HEX/OCT/BIN continue to show the raw bit pattern.
- Changing width or signedness clears the current entry and persists the choice.

The raw value is transported as a decimal string so 64-bit patterns survive the JavaScript boundary exactly.

### Operators

| Group | Operators |
| --- | --- |
| Bitwise | `AND`, `OR`, `XOR`, `NAND`, `NOR`, `NOT` |
| Shift / rotate | `SHL`, `SHR`, `SAR`, `ROL`, `ROR` |
| Arithmetic | `ADD`, `SUB`, `MUL`, `DIV`, `MOD`, `NEG` |

- Shifts by at least the word width produce 0 (or all ones for arithmetic right shift of a negative value).
- Rotates reduce the count modulo the width.
- `DIV` and `MOD` by zero are errors.
- Binary operators chain: pressing an operator with a pending operation evaluates the pending one first.
- `=` confirms the pending operation, or records the current value as history.
- `NEG` produces the two's complement; `NOT` complements within the word width.

### Keyboard

| Key | Action |
| --- | --- |
| `0`–`9`, `A`–`F` | Digits valid for the active base |
| `+`, `-`, `*`, `/`, `%` | `ADD`, `SUB`, `MUL`, `DIV`, `MOD` |
| `Enter` or `=` | Confirm / evaluate |
| `Backspace` | Delete digit; with an empty entry, reset the pending operation |
| `Escape` or `Delete` | Clear |

## Finance

Finance mode groups five calculators behind tabs. Inputs are normal form fields, so keyboard entry works as usual; `Enter` calculates the active tab.

### Keypad

A compact keypad below the fields writes into the field that was focused last (or the tab's first field). It evaluates arithmetic and percentages with the same rules as Normal mode: `250000 + 10%` commits `275000`, `250000 × 10%` commits `25000`, and a bare `%` divides by 100. On interest-rate fields the `%` key is disabled, because those values are already percentages (`6.5` means 6.5%/yr). The line above the keys previews the buffer; `=` applies it to the active field, `AC` discards it, and `Esc` clears the buffer and any validation error. With a pending buffer, the Calculate button or `Enter` applies it before calculating; with an empty buffer they calculate directly.

### Currency

A currency selector applies to all monetary output. Supported display currencies: USD, EUR, GBP, JPY, CAD, AUD, CHF, INR. Formatting uses `Intl.NumberFormat` with two fraction digits; the choice is persisted.

### Loan

Inputs: principal, annual interest rate (%), term in years, payments per year (12, 4 or 1).

Outputs:

- payment per period;
- number of periods;
- total interest and total paid;
- full amortization schedule with payment number, payment, principal, interest and remaining balance.

Formula: `P × r × (1+r)^n / ((1+r)^n − 1)`, with `r = annualRate / 100 / paymentsPerYear` and `n = years × paymentsPerYear`. A zero rate falls back to `P / n`. The last payment settles the balance exactly. Monetary values are rounded to two decimals.

Example: 250,000 at 6.5% over 30 years with 12 payments/year → 1,580.17 per month.

### Savings

Inputs: monthly deposit, annual rate (%), years.

Outputs: future value, total deposited, total interest. Deposits are end-of-month, `n = years × 12`, `r = annualRate / 100 / 12`, `FV = deposit × ((1+r)^n − 1) / r` (or `deposit × n` at zero rate).

### ROI

Inputs: initial value, final value. Output: `(final − initial) / |initial| × 100`, rounded to two decimals.

### CAGR

Inputs: initial value, final value, years. Output: `((final / initial)^(1/years) − 1) × 100`. Both values must be positive.

### Change

Inputs: starting value, ending value. Output: `(ending − starting) / |starting| × 100`. This is the signed percentage change; negative results mean a decrease.

### Validation

- Principal, deposits and values must be positive where required.
- Rates may be zero but not negative.
- Years must be positive whole numbers.
- A zero base for ROI or Change is an error.
- Errors appear below the form and do not enter history.

### History labels

Finance entries record a readable summary, for example:

```
Loan $250,000.00 @ 6.5% × 30y      $1,580.17/month
Savings $500.00/mo @ 5% × 15y      $133,644.47
ROI $10,000.00 → $16,000.00        60.00%
CAGR $10,000.00 → $16,000.00 / 5y  9.86%
```

## History

Every confirmed calculation is appended with its mode, expression, result and RFC 3339 timestamp. The newest entry appears first. The drawer supports deleting one entry with a slide-out animation or clearing everything; both actions persist and emit `history:changed`. The list is capped at 200 entries on disk. History survives restarts and is shared by all modes; the top-bar history button shows a dot when the list is non-empty.

## Memory

Normal and Scientific modes share an in-memory memory register:

- `MC` clears it.
- `MR` inserts the value into the current expression.
- `M+` / `M−` add or subtract the current preview/result.
- The bar displays `M <value>` when the register is non-zero.

Memory is intentionally not persisted across restarts, matching common calculator behaviour. The developer and finance surfaces do not expose it.
