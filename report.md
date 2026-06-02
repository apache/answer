# Access Control and Pseudonymous Content Support

## Project Overview

**Project:** Knowledge Sharing Platform
**Feature:** Pseudonymous Users & Content Visibility Control

## Objective

This fork extends the platform with mechanisms for user pseudonymity and content visibility control. The goal is to provide users with greater privacy when participating in discussions while ensuring that access to content can be restricted according to predefined visibility levels.

## Implemented Features

### 1. User Pseudonymity

Implemented a configurable anonymity system that allows users to participate without exposing their real identity.

#### Configuration Management

- Added anonymity configuration for users.
- Users can enable or disable anonymity mode at any time.
- Automatic creation of anonymity configuration during user registration.
- Automatic removal of anonymity configuration when a user account is permanently deleted.

#### Alias Generation

- Implemented fake name (alias) generation for anonymous users.
- Automatically creates alias records when anonymity mode is enabled and a user creates:
  - a question
  - an answer
  - a comment

#### Content Representation

- Replaced real user information with anonymous aliases when retrieving:
  - questions
  - answers
  - comments

- Replaced user avatars with anonymous representations.
- Administrators retain access to the original user information.

#### Access Restrictions

- Disabled navigation to anonymous user profiles.
- Implemented anonymization of notifications.
- Added support for viewing generated aliases by the alias owner

#### Cleanup Logic

- Removes all alias records associated with a question when that question is permanently deleted.

### 2. Content Visibility Control

Implemented configurable visibility levels for questions.

#### Visibility Levels

Added a visibility field represented by an enumeration:

- `public` – available to all users.
- `authenticated` – available only to authenticated users.
- `private` – available only to the author.

#### Access Enforcement

Implemented visibility validation across the application:

- Restricted access to questions according to their visibility level.
- Added permission checks before accessing question details.
- Prevented unauthorized navigation to restricted questions.
- Ensured question listings only contain content available to the current user.

## Backend Changes

### Data Model Updates

- Added anonymity configuration entity.
- Added alias storage.
- Added question visibility field and corresponding enum.

### Business Logic

- Alias generation and lifecycle management.
- Content anonymization during response mapping.
- Permission-based visibility filtering.
- Notification anonymization.

### Cleanup Operations

- Remove anonymity configuration on user deletion.
- Remove question aliases on question deletion.

## Result

The platform now supports:

- User-controlled anonymity.
- Automatic pseudonymous identity generation.
- Anonymous participation in discussions.
- Visibility-based access control for questions.
- Secure content filtering according to user permissions.
- Administrator access to original user information when required.
- Proper cleanup of anonymity-related data during entity removal.

These changes improve both privacy and access management while remaining fully integrated with the existing platform architecture.
