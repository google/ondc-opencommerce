#!/bin/bash
# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# key-rotation
key_rotation_services=('server_image_pusher_key_rotation')
for service in ${key_rotation_services[@]}; do
  echo "publish key-rotation $service"
  bazel run //docker/publish/key-rotation:$service --define DOCKER_REGISTRY="${1}" --define DOCKER_REPOSITORY="${2}"
done
