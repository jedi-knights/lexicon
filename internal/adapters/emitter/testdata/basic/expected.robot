*** Settings ***
Documentation    As a user, I want to page through search results so I can browse more than the first page of matches.
Test Setup    Background

*** Test Cases ***
Navigating to the next page
    [Tags]    search    pagination
    Given the user has performed a search that returns more than one page of results
    When the user clicks the "Next" pagination control
    Then the second page of results replaces the first page
    And the pagination control shows page 2 as active

Jumping to a specific page
    [Tags]    search    pagination
    [Template]    Jumping to a specific page
    2
    3

*** Keywords ***
Background
    Given the catalog has more than one page of results

Jumping to a specific page
    [Arguments]    ${page}
    Given the user has performed a search that returns more than one page of results
    When the user clicks the page ${page} control
    Then the ${page} page of results replaces the current page
