export namespace domain {
	
	export class Task {
	    id: string;
	    title: string;
	    short_title?: string;
	    description?: string;
	    status: string;
	    priority?: string;
	    due_date?: string;
	    order: number;
	    created_at: string;
	    updated_at: string;
	    completed_at?: string;
	    archived: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.short_title = source["short_title"];
	        this.description = source["description"];
	        this.status = source["status"];
	        this.priority = source["priority"];
	        this.due_date = source["due_date"];
	        this.order = source["order"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	        this.completed_at = source["completed_at"];
	        this.archived = source["archived"];
	    }
	}

}

