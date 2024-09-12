-- Define the character set for Base58
local ALPHABET_BASE58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

-- Convert the Base58 alphabet to a character table
local charset = {}
for i = 1, #ALPHABET_BASE58 do
    table.insert(charset, ALPHABET_BASE58:sub(i, i))
end

-- Function to generate random strings
local function randomString(length)
    math.randomseed(os.time() + math.random())
    if length > 0 then
        return randomString(length - 1) .. charset[math.random(1, #charset)]
    else
        return ""
    end
end

-- Open file for appending short URLs
local file = io.open("short_urls.txt", "a")

-- Function to log short URL to file
local function logShortUrl(short_url)
    file:write(short_url .. "\n")
    file:flush() -- Ensure the data is written to the file immediately
end

-- HTTP request function to be called for each request
request = function()
    -- Random string appended to URL
    local random_path = randomString(2)
    local random_domain = randomString(5)
    local body = '{"url": "https://subdomain-' .. random_domain .. '-something.io/library/react-' .. random_path .. '"}'

    -- Return the HTTP request to be sent
    return wrk.format("POST", "/encode", {["Content-Type"] = "application/json"}, body)
end

-- Response function to process each response
response = function(status, headers, body)
    if status == 200 then
        -- Parse the response body to extract the short_url
        local short_url = string.match(body, '"short_url"%s*:%s*"([^"]+)"')
        if short_url then
            -- Log the short URL to the file
            logShortUrl(short_url)
        end
    end
end

-- Cleanup function to close the file when the script is done
done = function(summary, latency, requests)
    if file then
        file:close()
    end
end
