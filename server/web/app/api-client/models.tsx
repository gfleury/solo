/**
 * 
 * @export
 * @interface Node
 */
export interface Node {
    ID?: number;
	peerid?:   string;
	hostname?:   string;
    os?:   string;
	arch?:   string;
	ip?:   string;
	version?:   string;
}

/**
 * 
 * @export
 * @interface Network
 */
export interface Network {
    ID?: number;
    name: string;
    cidr: string;
    nodes?: Array<Node>;
    user: User;
    linkedusers: Array<LinkedUser>;
}

/**
 * 
 * @export
 * @interface LinkedUser
 */
export interface LinkedUser {
    user?: User;
    permissions?: LinkedUser.PermissionsEnum;
}

/**
 * @export
 * @namespace LinkedUser
 */
export namespace LinkedUser {
    export enum PermissionsEnum {
        Write = 'write',
        Read = 'read',
        Owner = 'owner'
    }
}

/**
 * 
 * @export
 * @interface User
 */
export interface User {
    ID?: number;
    username?: string;
    email?: string;
    userStatus?: number;
    networks?: Array<Network>;
}