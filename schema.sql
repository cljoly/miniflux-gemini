-- The AGPLv3 License (AGPLv3)
--
-- Copyright (c) 2023 Clément Joly
--
-- This program is free software: you can redistribute it and/or modify
-- it under the terms of the GNU Affero General Public License as
-- published by the Free Software Foundation, either version 3 of the
-- License, or (at your option) any later version.
--
-- This program is distributed in the hope that it will be useful,
-- but WITHOUT ANY WARRANTY; without even the implied warranty of
-- MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
-- GNU Affero General Public License for more details.
--
-- You should have received a copy of the GNU Affero General Public License
-- along with this program.  If not, see <http://www.gnu.org/licenses/>.

CREATE TABLE IF NOT EXISTS Users (
	-- Public TLS certificate, to authenticate a user
	cert TEXT NOT NULL,
	-- Miniflux instance url
	instance TEXT NOT NULL,
	-- Miniflux API token
	token TEXT NOT NULL
) STRICT;
