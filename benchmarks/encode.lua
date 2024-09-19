

-- Define the character set for Base58
local ALPHABET_BASE58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
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

-- Open file for appending new short URLs
local file = io.open("short_urls_out.txt", "a")

-- Function to log new short URLs to file
local function logShortUrl(short_url)
    file:write(short_url .. "\n")
    file:flush() -- Ensure data is written to the file immediately
end

-- HTTP request function with 20% for /encode and 80% for /decode
local current_index = 1

request = function()
       local random_path = randomString(2)
        local random_domain = randomString(5)
        local body = '{"url": "https://subdomain-' .. random_domain .. '-something.io/library/react-' .. random_path .. '"}'
        return wrk.format("POST", "/api/encode", {["Content-Type"] = "application/json"}, body)
end

-- Response function to log short URL from /encode response
response = function(status, headers, body)
    if status == 200 then
        -- Extract the short URL from the /encode response
        local short_url = string.match(body, '"shortUrl"%s*:%s*"([^"]+)"')
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
