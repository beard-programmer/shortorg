-- -- Read all short URLs from the file into a table
-- local short_urls = {}
--
-- local function loadShortUrlsFromFile(filename)
--     local file = io.open(filename, "r")
--     if not file then
--         error("Could not open file " .. filename)
--     end
--
--     for line in file:lines() do
--         table.insert(short_urls, line)
--     end
--     file:close()
-- end
--
-- -- Load the short URLs from the file
-- loadShortUrlsFromFile("short_urls.txt")
--
-- -- Index for keeping track of the current short URL
-- local current_index = 1

-- HTTP request function to be called for each request
request = function()
--     if current_index > #short_urls then
--         current_index = 1 -- Reset the index if we reach the end of the list
--     end

--     -- Use the short URL from the list
--     local short_url = short_urls[current_index]
--     current_index = current_index + 1

    -- Create the HTTP request body
    local body = '{"shortUrl": "https://shortl.org/2E1Moa"}'

    -- Return the HTTP request to be sent
    return wrk.format("POST", "/api/decode", {["Content-Type"] = "application/json"}, body)
end
