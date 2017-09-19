const storage = sessionStorage
const prefix = "ovpm_auth_"
export function GetAuthToken() {
    let key = storage.getItem("authKey")
    if (!key) {
        return ""
    }
    return key
}

export function SetAuthToken(token) {
    return storage.setItem("authKey", token)
}

export function ClearAuthToken() {
    storage.removeItem("authKey")
}

export function IsAuthenticated() {
    if (GetAuthToken()) {
        return true
    }
    return false
}

export function SetItem(key, value) {
    if (key.constructor === String) {
        return storage.setItem(JSON.stringify(prefix+key), JSON.stringify(value))
    }
    throw "key should be a string: " + key
}

export function GetItem(key) {
    if (key.constructor === String) {
        return JSON.parse(storage.getItem(JSON.stringify(prefix+key)))
    }
    throw "key should be a string: " + key
}
