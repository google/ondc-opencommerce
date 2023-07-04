-- Copyright 2023 Google LLC
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

CREATE TABLE Registration(
  RegistrationID STRING(35) NOT NULL,
  ONDCEnvironment STRING(30) NOT NULL,
  ONDCRegistryURL STRING(255) NOT NULL,
  EntityName STRING(255) NOT NULL,
  BusinessAddress STRING(255) NOT NULL,
  GSTDetails STRING(255) NOT NULL,
  PANNo STRING(25) NOT NULL,
  AddressSignatory STRING(255) NOT NULL,
  EmailID STRING(255) NOT NULL,
  MobileNumber STRING(15) NOT NULL,
  DomainsEnabled STRING(255) NOT NULL,
  AppType STRING(255) NOT NULL,
  SubscriberID STRING(10) NOT NULL,
  SubscriberURL STRING(255) NOT NULL,
  CreationTime TIMESTAMP
  DEFAULT(CURRENT_TIMESTAMP),
  LastModifiedTime TIMESTAMP
  OPTIONS (allow_commit_timestamp = TRUE),
  CreationBy STRING(255),
  LastModifiedBy STRING(255),
  AdditionalData JSON,)
  PRIMARY KEY(RegistrationID)
