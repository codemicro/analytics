{
    const modalLoadURL = "{{#url}}"
    const modalTargetID = "{{#id}}"

    // Create edit button in page header
    const pageHeader = document.querySelector("div.page-header")
    
    const editElem = document.createElement("a")
    editElem.href = "#"
    editElem.innerText = "(edit)"
    editElem.setAttribute("hx-get", modalLoadURL)
    editElem.setAttribute("hx-target", "#" + modalTargetID)

    pageHeader.appendChild(editElem)

    // Create div to become the modal target
    const modalTarget = document.createElement("div")
    modalTarget.id = modalTargetID
    document.body.appendChild(modalTarget)

    // Load HTMX
    // TODO: Bundle this and don't rely on a CDN somehow?
    const htmxImport = document.createElement("script")
    htmxImport.src = "https://unpkg.com/htmx.org@1.8.6"
    document.body.appendChild(htmxImport)
}