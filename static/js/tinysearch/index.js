const searchForm = document.getElementById("search-form");
const searchInputEl = document.getElementById("search-input");
const resultsDiv = document.getElementById("results");
const statusEl = document.getElementById("status");

let wasmModule = null;
let memory = null;
let searchFunction = null;
let freeFunction = null;

async function initializeSearch() {
  try {
    // Fetch from the WASM module and assemble the response
    const response = await fetch("./js/tinysearch/tinysearch_engine.wasm");
    if (!response.ok) {
      throw new Error(
        `Failed to fetch WASM: ${response.status} ${response.statusText}`
      );
    }
    const wasmBytes = await response.arrayBuffer();
    const module = await WebAssembly.instantiate(wasmBytes);

    // Verify the returned module includes the necessary exports
    wasmModule = module.instance;
    memory = wasmModule.exports.memory;
    searchFunction = wasmModule.exports.search;
    freeFunction = wasmModule.exports.free_search_result;
    if (!searchFunction || !freeFunction || !memory) {
      throw new Error("Required WASM exports not found");
    }

    // Prepare the form, because the search engine ready for queries
    searchForm.addEventListener("submit", performSearch);
    searchForm.disabled = false;
    statusEl.textContent = "Search engine ready! Try entering a query above.";
    searchInputEl.focus();
  } catch (error) {
    console.error("Failed to initialize search:", error);
    statusEl.innerHTML = `<div class="error">Failed to initialize search: ${error.message}</div>`;
  }
}

function stringToWasmPtr(str) {
  const bytes = new TextEncoder().encode(str + "\0");
  const ptr = wasmModule.exports?.__wbindgen_malloc?.(bytes.length);
  if (!ptr) {
    // Fallback: write to a known memory location
    const memoryArray = new Uint8Array(memory.buffer);
    const startOffset = 1024; // Use a safe offset
    memoryArray.set(bytes, startOffset);
    return startOffset;
  }
  new Uint8Array(memory.buffer, ptr, bytes.length).set(bytes);
  return ptr;
}

function wasmPtrToString(ptr) {
  if (ptr === 0) return null;

  const memoryArray = new Uint8Array(memory.buffer);
  let length = 0;

  // Find the null terminator
  while (memoryArray[ptr + length] !== 0) {
    length++;
    if (length > 1000000) break; // Safety limit
  }

  const bytes = memoryArray.slice(ptr, ptr + length);
  return new TextDecoder().decode(bytes);
}

function performSearch(evt) {
  evt.preventDefault();

  const query = searchInputEl.value.trim();

  if (!wasmModule) {
    statusEl.innerHTML =
      '<div class="error">Search engine not initialized</div>';
    return;
  }

  if (!query) {
    resultsDiv.innerHTML = "";
    statusEl.textContent = "Enter a search query to see results.";
    return;
  }

  try {
    const startTime = performance.now();

    // Allocate memory for query string
    let queryPtr;
    try {
      queryPtr = stringToWasmPtr(query);
    } catch (e) {
      // Fallback: use a fixed memory location
      const queryBytes = new TextEncoder().encode(query + "\0");
      const memoryArray = new Uint8Array(memory.buffer);
      queryPtr = 1024; // Fixed offset
      memoryArray.set(queryBytes, queryPtr);
    }

    // Call search function
    const resultPtr = searchFunction(queryPtr, 10);

    // Free query memory if we allocated it
    if (wasmModule.exports.__wbindgen_free && queryPtr > 1024) {
      wasmModule.exports.__wbindgen_free(queryPtr, query.length + 1);
    }

    const endTime = performance.now();
    const searchTime = (endTime - startTime).toFixed(3);

    if (resultPtr === 0) {
      resultsDiv.innerHTML = '<div class="no-results">No results found</div>';
      statusEl.textContent = `No results found for "${query}" (${searchTime}ms)`;
      return;
    }

    // Read result string
    const resultString = wasmPtrToString(resultPtr);

    // Free result memory
    freeFunction(resultPtr);

    if (!resultString) {
      resultsDiv.innerHTML = '<div class="no-results">No results found</div>';
      statusEl.textContent = `No results found for "${query}" (${searchTime}ms)`;
      return;
    }

    const results = JSON.parse(resultString);

    if (results.length === 0) {
      resultsDiv.innerHTML = '<div class="no-results">No results found</div>';
      statusEl.textContent = `No results found for "${query}" (${searchTime}ms)`;
    } else {
      const resultHTML = results
        .map(
          (result) => `
                        <div class="result-item">
                            <a href="${result.url}" class="result-title" target="_blank">${result.title}</a>
                        </div>
                    `
        )
        .join("");

      resultsDiv.innerHTML = resultHTML;
      statusEl.textContent = `Found ${results.length} result${
        results.length !== 1 ? "s" : ""
      } for "${query}" (${searchTime}ms)`;
    }
  } catch (error) {
    console.error("Search error:", error);
    resultsDiv.innerHTML =
      '<div class="error">Search failed. Check console for details.</div>';
    statusEl.innerHTML = `<div class="error">Search failed: ${error.message}</div>`;
  }
}

// Initialize search engine
document.getElementById("search-form").disabled = true;
initializeSearch();
