/**
 * 
 * @export
 * @interface Network
 */
export interface Network {
    /**
     * 
     * @type {number}
     * @memberof Network
     */
    ID?: number;
    /**
     * 
     * @type {string}
     * @memberof Network
     */
    name: string;
    cidr: string;
}

/**
 * 
 * @export
 * @interface LinkedUser
 */
export interface LinkedUser {
    /**
     * 
     * @type {User}
     * @memberof LinkedUser
     */
    user?: User;
    /**
     * network status
     * @type {string}
     * @memberof LinkedUser
     */
    permissions?: LinkedUser.PermissionsEnum;
}

/**
 * @export
 * @namespace LinkedUser
 */
export namespace LinkedUser {
    /**
     * @export
     * @enum {string}
     */
    export enum PermissionsEnum {
        Write = 'write',
        Read = 'read',
        Owner = 'owner'
    }
}
/**
 * 
 * @export
 * @interface Provider
 */
export interface Provider {
    /**
     * 
     * @type {number}
     * @memberof Provider
     */
    ID?: number;
    /**
     * 
     * @type {string}
     * @memberof Provider
     */
    name?: string;
    authenticationType?: string;
}
/**
 * 
 * @export
 * @interface Session
 */
export interface Session {
    /**
     * 
     * @type {number}
     * @memberof Session
     */
    ID?: number;
    /**
     * 
     * @type {number}
     * @memberof Session
     */
    networkID?: number;
    /**
     * 
     * @type {string}
     * @memberof Session
     */
    sessionData?: string;
}
/**
 * 
 * @export
 * @interface Tag
 */
export interface Tag {
    /**
     * 
     * @type {number}
     * @memberof Tag
     */
    ID?: number;
    /**
     * 
     * @type {string}
     * @memberof Tag
     */
    name?: string;
}
/**
 * 
 * @export
 * @interface User
 */
export interface User {
    /**
     * 
     * @type {number}
     * @memberof User
     */
    ID?: number;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    username?: string;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    firstName?: string;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    lastName?: string;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    email?: string;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    password?: string;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    phone?: string;
    /**
     * User Status
     * @type {number}
     * @memberof User
     */
    userStatus?: number;
    /**
     * Networks that belong to user
     * @type {Array<Network>}
     * @memberof User
     */
    networks?: Array<Network>;
    /**
     * Users that are linked to this user
     * @type {Array<LinkedUser>}
     * @memberof User
     */
    linkedUsers?: Array<LinkedUser>;
    /**
     * User has confirmed email
     * @type {boolean}
     * @memberof User
     */
    verified?: boolean;
    avatar?: string;
}