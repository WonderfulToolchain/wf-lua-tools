#!/opt/wonderful/bin/wf-lua
-- SPDX-License-Identifier: MIT
-- SPDX-FileContributor: Adrian "asie" Siekierka, 2024

commands = {}
commands.migrate = require('wf.internal.tool.config.migrate')
commands.repo = require('wf.internal.tool.config.repo')

require('wf.internal.tool')

