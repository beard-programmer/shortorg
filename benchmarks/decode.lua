-- Read all short URLs from the file into a table
local short_urls = {}

local function loadShortUrlsFromFile(filename)
    local file = io.open(filename, "r")
    if not file then
        error("Could not open file " .. filename)
    end

    for line in file:lines() do
        table.insert(short_urls, line)
    end
    file:close()
end

-- Load the short URLs from the file
loadShortUrlsFromFile("short_urls_out.txt")

math.randomseed(os.time())


-- HTTP request function to be called for each request
request = function()


    local random_index = math.random(1, #short_urls)
    local short_url = short_urls[random_index]

    -- Create the HTTP request body
    local body = '{"shortUrl": "' .. short_url .. '"}'

    -- Return the HTTP request to be sent
    return wrk.format("POST", "/api/resolve-link", {["Content-Type"] = "application/json"}, body)
end
