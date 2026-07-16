Feature: Neutral dialect example

  Background:
    Given the catalog has more than one page of results

  Scenario: Neutral steps translate to Gherkin
    Given the user has performed a search that returns more than one page of results
    When the user clicks the "Next" pagination control
    Then the second page of results replaces the first page
    And the pagination control shows page 2 as active
