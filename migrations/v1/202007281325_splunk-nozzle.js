exports.migrate = function(input) {
    // Append text to a string
    if (input.properties['.forms.{advanced}.properties.{add_app_info}']['value']=='true') {
        input.properties['.forms.{advanced}.properties.{add_app_info}']['value'] = "AppName,OrgName,OrgGuid,SpaceName,SpaceGuid"
    } else {
        input.properties['.forms.{advanced}.properties.{add_app_info}']['value'] = ""
    }
    return input;
};
