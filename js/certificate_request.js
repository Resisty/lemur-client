(function( $ ) {
    'use strict';
    function getCookie(cname) {
        var name = cname + "=";
        var decodedCookie = decodeURIComponent(document.cookie);
        var ca = decodedCookie.split(';');
        for(var i = 0; i <ca.length; i++) {
            var c = ca[i];
            while (c.charAt(0) == ' ') {
                c = c.substring(1);
            }
            if (c.indexOf(name) == 0) {
                return c.substring(name.length, c.length);
            }
        }
        return "";
    }
    function clearTextAreas() {
        $('#authority').val('')
        $('#commonName').val('')
        $('#owner').val('')
        $('#validityStart').val('')
        $('#validityEnd').val('')
    }
   function emptyInputs() {
        var vals = [$('#authority').val(),
                    $('#commonName').val(),
                    $('#owner').val(),
                    $('#validityStart').val(),
                    $('#validityEnd').val()]
        return jQuery.grep(vals, function(n) {
            return n == '';
        }).length > 0;
    }
    function makeCertPanels(dict) {
        var div = $('#certificate-data');
        div.empty();
        var chainPanel = $("<div>",
                           {class: 'panel panel-default'});
        chainPanel.appendTo(div);
        var chainPanelHead = $("<div>",
                               {class: "panel-heading"});
        chainPanelHead.appendTo(chainPanel);
        var chainH3 = $("<h3>",
                        {class: 'panel-title',
                         text: 'Certificate Chain'});
        chainH3.appendTo(chainPanelHead);
    var chainBody = $("<div>",
                  {class: 'panel-body'});
    chainBody.appendTo(chainPanelHead);
    var chainPre = $("<pre>",
                 {text: dict.chain});
    chainPre.appendTo(chainBody);

        var pubCertPanel = $("<div>",
                             {class: 'panel panel-default'});
        pubCertPanel.appendTo(div);
        var pubCertPanelHead = $("<div>",
                                 {class: "panel-heading"});
        pubCertPanelHead.appendTo(pubCertPanel);
        var pubCertH3 = $("<h3>",
                          {class: 'panel-title',
                           text: 'Public Certificate'});
        pubCertH3.appendTo(pubCertPanelHead);
    var pubCertBody = $("<div>",
                            {class: 'panel-body'});
    pubCertBody.appendTo(pubCertPanelHead);
    var pubCertPre = $("<pre>",
                 {text: dict.pubcert});
    pubCertPre.appendTo(pubCertBody);

        var privKeyPanel = $("<div>",
                             {class: 'panel panel-default'});
        privKeyPanel.appendTo(div);
        var privKeyPanelHead = $("<div>",
                                 {class: "panel-heading"});
        privKeyPanelHead.appendTo(privKeyPanel);
        var privKeyH3 = $("<h3>",
                          {class: 'panel-title',
                           text: 'Private Key'});
        privKeyH3.appendTo(privKeyPanelHead);
    var privKeyBody = $("<div>",
                    {class: 'panel-body'});
    privKeyBody.appendTo(privKeyPanelHead);
    var privKeyPre = $("<pre>",
                 {text: dict.privatekey});
    privKeyPre.appendTo(privKeyBody);
    }
    $(document).ready(function() {
    $('#clear').click( function() {
        clearCertPanels()
    });
    $('#submit').click( function(event) {
        if ( emptyInputs() ) {
            alert('This form requires Authority, CommonName (you), Owner (email), StartDate, EndDate and RBAC Group!');
        } else {
        var data = {};
        var verb = 'POST';
        var url = '/v1/createcert';
        data['authority']     = $('#authority').val();
        data['commonName']    = $('#commonName').val();
        data['owner']         = $('#owner').val();
        data['validityStart'] = $('#validityStart').val();
        data['validityEnd']   = $('#validityEnd').val();
        var result = $.ajax({
            type: verb,
            url: url,
            dataType: 'json',
            data: JSON.stringify(data),
            headers: {'Authorization': getCookie('auth')},
            success: function(data) {
            makeCertPanels(data);
            clearTextAreas();
            },
            error: function(xhr, ajaxOptions, thrownError) { 
            console.log(xhr.responseText);
            alert(xhr.responseText);
            },
            contentType: 'application/json',
        });
        }
    });
    });
})( jQuery );

