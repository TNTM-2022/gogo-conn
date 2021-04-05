"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
function default_1(app) {
    return new GateHandler(app);
}
exports.default = default_1;
class GateHandler {
    constructor(app) {
        this.app = app;
    }
    async test(msg, session) {
        return {
            code: 200,
            name: 'test',
            age: 10
        };
    }
}
exports.GateHandler = GateHandler;
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiZ2F0ZUhhbmRsZXIuanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyIuLi8uLi8uLi8uLi8uLi9hcHAvc2VydmVycy9nYXRlL2hhbmRsZXIvZ2F0ZUhhbmRsZXIudHMiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6Ijs7QUFHQSxtQkFBeUIsR0FBZ0I7SUFDckMsT0FBTyxJQUFJLFdBQVcsQ0FBQyxHQUFHLENBQUMsQ0FBQztBQUNoQyxDQUFDO0FBRkQsNEJBRUM7QUFFRCxNQUFhLFdBQVc7SUFDcEIsWUFBb0IsR0FBZ0I7UUFBaEIsUUFBRyxHQUFILEdBQUcsQ0FBYTtJQUNwQyxDQUFDO0lBRUQsS0FBSyxDQUFDLElBQUksQ0FBRSxHQUFrQixFQUFFLE9BQXVCO1FBQ25ELE9BQU87WUFDSCxJQUFJLEVBQUUsR0FBRztZQUNULElBQUksRUFBRSxNQUFNO1lBQ1osR0FBRyxFQUFFLEVBQUU7U0FDVixDQUFBO0lBQ0wsQ0FBQztDQThCSjtBQXhDRCxrQ0F3Q0MifQ==