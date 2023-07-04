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

CREATE TABLE Transaction(
  TransactionID STRING(36) NOT NULL,
  TransactionType INT64 NOT NULL,
  TransactionAPI INT64 NOT NULL,
  MessageID STRING(36) NOT NULL,
  RequestID STRING(36) DEFAULT (GENERATE_UUID()),
  Payload JSON NOT NULL,
  ProviderID STRING(255) NOT NULL,
  MessageStatus STRING(5),
  ErrorType STRING(36),
  ErrorCode STRING(255),
  ErrorPath STRING(MAX),
  ErrorMessage STRING(MAX),
  ReqReceivedTime TIMESTAMP,
  AdditionalData JSON,)
  PRIMARY KEY(TransactionID, TransactionType, MessageID, RequestID)
