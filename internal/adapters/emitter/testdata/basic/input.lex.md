---
id: SEARCH-042
owner: product
status: draft
---

# Feature: Search results pagination

As a user, I want to page through search results so I can browse more than the first page of matches.

## Background

- **Given** the catalog has more than one page of results

@search @pagination
## Scenario: Navigating to the next page

- **Given** the user has performed a search that returns more than one page of results
- **When** the user clicks the "Next" pagination control
- **Then** the second page of results replaces the first page
- **And** the pagination control shows page 2 as active

@search @pagination
## Scenario: Jumping to a specific page

- **Given** the user has performed a search that returns more than one page of results
- **When** the user clicks the page `<page>` control
- **Then** the `<page>` page of results replaces the current page

| page |
| ---- |
| 2    |
| 3    |
