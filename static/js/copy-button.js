document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll("pre code").forEach((codeBlock) => {
    // Create copy button element
    const button = document.createElement("button");
    button.className = "btn btn-copy";
    button.type = "button";
    button.textContent = "Copy";
    button.setAttribute("aria-label", "Copy code to clipboard");

    button.addEventListener("click", async () => {
      // Remove line numbers
      let text = codeBlock.textContent;
      text = text
        .split("\n")
        .map((line) => line.replace(/^\s*\d+/, ""))
        .join("\n");

      // Copy text to clipboard
      await navigator.clipboard.writeText(text);
      button.textContent = "Copied!";
      setTimeout(() => (button.textContent = "Copy"), 2000);
    });

    codeBlock.parentElement.insertBefore(button, codeBlock);
  });
});
