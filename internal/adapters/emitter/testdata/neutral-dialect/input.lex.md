# Feature: Neutral dialect example

## Background

- **Precondition** the catalog has more than one page of results

## Scenario: Neutral steps translate to Gherkin

- **Precondition** the user has performed a search that returns more than one page of results
- **Action** the user clicks the "Next" pagination control
- **Outcome** the second page of results replaces the first page
- **And** the pagination control shows page 2 as active
